using System;
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
        }

        private string SanitizeFilename(string filename)
        {
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
