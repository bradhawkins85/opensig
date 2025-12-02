using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using OpenSig.Agent.Models;

namespace OpenSig.Agent.Services
{
    public class SignatureWriter
    {
        private readonly string _signaturesDirectory;

        public SignatureWriter()
        {
            string appData = Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData);
            _signaturesDirectory = Path.Combine(appData, "Microsoft", "Signatures");
        }

        public SignatureWriter(string signaturesDirectory)
        {
            _signaturesDirectory = signaturesDirectory;
        }

        public void WriteSignatures(AgentTemplateResponse response)
        {
            // Create signatures directory if it doesn't exist
            Directory.CreateDirectory(_signaturesDirectory);

            Logger.Log($"Writing signatures to {_signaturesDirectory}");

            foreach (var template in response.Templates)
            {
                WriteSignature(template);
            }

            Logger.Log($"Successfully wrote {response.Templates.Count} signature(s)");

            // Set default signatures in Outlook if enabled
            if (response.SetDefaultSignatures && response.Templates.Count > 0)
            {
                Logger.Log("Feature flag 'set_default_signatures' is enabled. Configuring Outlook defaults...");
                
                var registryManager = new OutlookRegistryManager();
                
                // Check if roaming signatures are enabled
                if (registryManager.IsRoamingSignaturesEnabled())
                {
                    Logger.LogWarning("Roaming signatures are enabled. Skipping default signature configuration to avoid conflicts.");
                }
                else
                {
                    // Use the first template as the default signature
                    var firstTemplate = response.Templates[0];
                    string safeName = SanitizeFilename(firstTemplate.Name);
                    
                    if (registryManager.SetDefaultSignatures(safeName))
                    {
                        Logger.Log($"Successfully configured '{safeName}' as the default signature for new emails and replies");
                    }
                    else
                    {
                        Logger.LogWarning("Could not configure default Outlook signatures. See previous warnings for details.");
                    }
                }
            }
            else if (response.SetDefaultSignatures && response.Templates.Count == 0)
            {
                Logger.LogWarning("Cannot set default signatures: No templates available");
            }
        }

        private void WriteSignature(RenderedTemplate template)
        {
            // Sanitize the template name for use as a filename
            string safeName = SanitizeFilename(template.Name);

            // Write HTML file (.htm)
            string htmlPath = Path.Combine(_signaturesDirectory, $"{safeName}.htm");
            File.WriteAllText(htmlPath, template.HtmlContent, Encoding.UTF8);
            Logger.Log($"Wrote HTML signature: {htmlPath}");

            // Write RTF file (.rtf)
            string rtfPath = Path.Combine(_signaturesDirectory, $"{safeName}.rtf");
            File.WriteAllText(rtfPath, template.RtfContent, Encoding.UTF8);
            Logger.Log($"Wrote RTF signature: {rtfPath}");

            // Write plain text file (.txt)
            string txtPath = Path.Combine(_signaturesDirectory, $"{safeName}.txt");
            File.WriteAllText(txtPath, template.TextContent, Encoding.UTF8);
            Logger.Log($"Wrote TXT signature: {txtPath}");

            // Write assets to the _files directory (Outlook convention)
            WriteAssets(template, safeName);
        }

        private void WriteAssets(RenderedTemplate template, string safeName)
        {
            if (template.Assets == null || template.Assets.Count == 0)
            {
                return;
            }

            // Create the assets directory following Outlook's convention: {SignatureName}_files/
            string assetsDirectory = Path.Combine(_signaturesDirectory, $"{safeName}_files");
            Directory.CreateDirectory(assetsDirectory);

            Logger.Log($"Writing {template.Assets.Count} asset(s) to {assetsDirectory}");

            foreach (var asset in template.Assets)
            {
                WriteAsset(assetsDirectory, asset);
            }
        }

        // Maximum asset file size (10 MB) to prevent memory exhaustion
        private const int MaxAssetSizeBytes = 10 * 1024 * 1024;
        
        // Allowed image extensions for signature assets
        private static readonly HashSet<string> AllowedAssetExtensions = new(StringComparer.OrdinalIgnoreCase)
        {
            ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".webp"
        };

        private void WriteAsset(string assetsDirectory, TemplateAsset asset)
        {
            try
            {
                // Validate the base64 data size before decoding to prevent memory exhaustion
                // Base64 encoding increases size by approximately 4/3, so calculate approximate decoded size
                if (asset.Data.Length > MaxAssetSizeBytes * 4 / 3)
                {
                    Logger.LogWarning($"Skipping asset '{asset.Filename}': Exceeds maximum allowed size of {MaxAssetSizeBytes / 1024 / 1024} MB");
                    return;
                }

                // Extract just the filename, removing any path components
                string rawFilename = Path.GetFileName(asset.Filename);
                
                // Validate file extension is an allowed image type
                string extension = Path.GetExtension(rawFilename);
                if (string.IsNullOrEmpty(extension) || !AllowedAssetExtensions.Contains(extension))
                {
                    Logger.LogWarning($"Skipping asset '{asset.Filename}': Extension '{extension}' is not an allowed image type");
                    return;
                }

                // Sanitize filename to prevent path traversal
                string safeFilename = SanitizeFilename(rawFilename);
                if (string.IsNullOrEmpty(safeFilename))
                {
                    Logger.LogWarning($"Skipping asset with invalid filename: {asset.Filename}");
                    return;
                }

                // Additional path traversal protection: verify the final path is within the assets directory
                string assetPath = Path.Combine(assetsDirectory, safeFilename);
                string fullAssetPath = Path.GetFullPath(assetPath);
                string fullAssetsDirectory = Path.GetFullPath(assetsDirectory);
                
                if (!fullAssetPath.StartsWith(fullAssetsDirectory, StringComparison.OrdinalIgnoreCase))
                {
                    Logger.LogWarning($"Skipping asset '{asset.Filename}': Path traversal attempt detected");
                    return;
                }

                // Decode base64 data and write to file
                byte[] assetData = Convert.FromBase64String(asset.Data);
                
                // Validate decoded size as an additional check
                if (assetData.Length > MaxAssetSizeBytes)
                {
                    Logger.LogWarning($"Skipping asset '{asset.Filename}': Decoded size exceeds maximum allowed size of {MaxAssetSizeBytes / 1024 / 1024} MB");
                    return;
                }

                File.WriteAllBytes(assetPath, assetData);

                Logger.Log($"Wrote asset: {assetPath} ({asset.ContentType}, {assetData.Length} bytes)");
            }
            catch (FormatException ex)
            {
                Logger.LogWarning($"Failed to decode asset '{asset.Filename}': Invalid base64 data - {ex.Message}");
            }
            catch (Exception ex)
            {
                Logger.LogError($"Failed to write asset '{asset.Filename}': {ex.Message}");
            }
        }

        private string SanitizeFilename(string filename)
        {
            if (string.IsNullOrEmpty(filename))
            {
                return string.Empty;
            }

            // Remove invalid filename characters
            foreach (char c in Path.GetInvalidFileNameChars())
            {
                filename = filename.Replace(c, '_');
            }
            
            // Also replace some additional problematic characters
            filename = filename.Replace(' ', '_').Replace(':', '_').Replace('/', '_').Replace('\\', '_');
            
            return filename;
        }
    }
}
