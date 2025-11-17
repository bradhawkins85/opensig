using System;
using System.Net.Http;
using System.Text.Json;
using System.Threading.Tasks;
using OpenSig.Agent.Models;

namespace OpenSig.Agent.Services
{
    public class ApiClient
    {
        private readonly HttpClient _httpClient;
        private readonly string _apiBaseUrl;

        public ApiClient(string apiBaseUrl)
        {
            _apiBaseUrl = apiBaseUrl.TrimEnd('/');
            _httpClient = new HttpClient
            {
                Timeout = TimeSpan.FromSeconds(30)
            };
        }

        public async Task<AgentTemplateResponse?> GetTemplatesAsync(string? userEmail = null, string? userId = null, string? tenantId = null)
        {
            try
            {
                var request = new HttpRequestMessage(HttpMethod.Get, $"{_apiBaseUrl}/v1/agent/templates");
                
                // Add headers for user context (in production, these would be from authentication)
                if (!string.IsNullOrEmpty(userEmail))
                {
                    request.Headers.Add("X-User-Email", userEmail);
                }
                if (!string.IsNullOrEmpty(userId))
                {
                    request.Headers.Add("X-User-ID", userId);
                }
                if (!string.IsNullOrEmpty(tenantId))
                {
                    request.Headers.Add("X-Tenant-ID", tenantId);
                }

                Logger.Log($"Fetching templates from {_apiBaseUrl}/v1/agent/templates");
                
                var response = await _httpClient.SendAsync(request);
                response.EnsureSuccessStatusCode();

                var content = await response.Content.ReadAsStringAsync();
                var result = JsonSerializer.Deserialize<AgentTemplateResponse>(content);

                Logger.Log($"Received {result?.Templates?.Count ?? 0} templates from API");
                
                return result;
            }
            catch (HttpRequestException ex)
            {
                Logger.LogError($"HTTP error fetching templates: {ex.Message}");
                throw;
            }
            catch (Exception ex)
            {
                Logger.LogError($"Error fetching templates: {ex.Message}");
                throw;
            }
        }
    }
}
