using System;
using System.IO;
using System.Text;

namespace OpenSig.Agent
{
    internal class Program
    {
        static int Main(string[] args)
        {
            try
            {
                string appData = Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData);
                string sigDir = Path.Combine(appData, "Microsoft", "Signatures");
                Directory.CreateDirectory(sigDir);

                // Minimal example signature
                string name = "OpenSigSignature";
                string htmlPath = Path.Combine(sigDir, $"{name}.htm");
                string rtfPath  = Path.Combine(sigDir, $"{name}.rtf");
                string txtPath  = Path.Combine(sigDir, $"{name}.txt");

                File.WriteAllText(htmlPath,
                    "<div style='font-family:Segoe UI,Arial,sans-serif'><b>OpenSig (preview)</b><br/>Local Windows Agent</div>",
                    Encoding.UTF8);
                File.WriteAllText(rtfPath, @"{\\rtf1\\ansi OpenSig (preview) - Local Windows Agent}");
                File.WriteAllText(txtPath, "OpenSig (preview) - Local Windows Agent");

                Console.WriteLine($"Wrote signature to: {sigDir}");
                Console.WriteLine("NOTE: Setting Outlook default signatures is not implemented in this stub.");

                return 0;
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine(ex.ToString());
                return 1;
            }
        }
    }
}
