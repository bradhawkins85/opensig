using System;
using System.Threading.Tasks;
using OpenSig.Agent.Services;

namespace OpenSig.Agent
{
    internal class Program
    {
        static async Task<int> Main(string[] args)
        {
            Logger.Log("OpenSig Windows Agent starting...");

            try
            {
                // Get API URL from environment variable or use default
                string apiUrl = Environment.GetEnvironmentVariable("OPENSIG_API_URL") ?? "http://localhost:8080";
                
                // Get user info from environment or use defaults for testing
                string? userEmail = Environment.GetEnvironmentVariable("OPENSIG_USER_EMAIL");
                string? userId = Environment.GetEnvironmentVariable("OPENSIG_USER_ID");
                string? tenantId = Environment.GetEnvironmentVariable("OPENSIG_TENANT_ID");

                Logger.Log($"Connecting to API: {apiUrl}");

                // Create API client
                var apiClient = new ApiClient(apiUrl);

                // Fetch templates from API
                var response = await apiClient.GetTemplatesAsync(userEmail, userId, tenantId);

                if (response == null || response.Templates.Count == 0)
                {
                    Logger.LogWarning("No templates received from API");
                    return 1;
                }

                Logger.Log($"Authenticated as: {response.UserEmail}");

                // Write signatures to disk
                var writer = new SignatureWriter();
                writer.WriteSignatures(response);

                Logger.Log("Signature sync completed successfully");

                return 0;
            }
            catch (Exception ex)
            {
                Logger.LogError($"Fatal error: {ex.Message}");
                Logger.LogError(ex.ToString());
                return 1;
            }
        }
    }
}
