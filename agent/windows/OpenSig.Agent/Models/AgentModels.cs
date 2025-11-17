using System;
using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace OpenSig.Agent.Models
{
    public class AgentTemplateResponse
    {
        [JsonPropertyName("templates")]
        public List<RenderedTemplate> Templates { get; set; } = new();

        [JsonPropertyName("user_email")]
        public string UserEmail { get; set; } = string.Empty;

        [JsonPropertyName("user_id")]
        public string UserId { get; set; } = string.Empty;

        [JsonPropertyName("set_default_signatures")]
        public bool SetDefaultSignatures { get; set; } = false;
    }

    public class RenderedTemplate
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = string.Empty;

        [JsonPropertyName("name")]
        public string Name { get; set; } = string.Empty;

        [JsonPropertyName("html_content")]
        public string HtmlContent { get; set; } = string.Empty;

        [JsonPropertyName("rtf_content")]
        public string RtfContent { get; set; } = string.Empty;

        [JsonPropertyName("text_content")]
        public string TextContent { get; set; } = string.Empty;
    }
}
