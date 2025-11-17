package renderer

import (
	"testing"
)

func TestRender_SimplePlaceholder(t *testing.T) {
	template := "Hello, {{name}}!"
	data := map[string]string{
		"name": "John Doe",
	}
	
	result := Render(template, data)
	expected := "Hello, John Doe!"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_MultiplePlaceholders(t *testing.T) {
	template := "{{firstName}} {{lastName}} - {{title}}"
	data := map[string]string{
		"firstName": "Jane",
		"lastName":  "Smith",
		"title":     "Senior Engineer",
	}
	
	result := Render(template, data)
	expected := "Jane Smith - Senior Engineer"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_MissingPlaceholder(t *testing.T) {
	template := "Hello, {{name}}!"
	data := map[string]string{}
	
	result := Render(template, data)
	expected := "Hello, {{name}}!" // Placeholder should remain
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_ConditionalTrue(t *testing.T) {
	template := "Name: {{name}}{{#if title}}\nTitle: {{title}}{{/if}}"
	data := map[string]string{
		"name":  "John Doe",
		"title": "Manager",
	}
	
	result := Render(template, data)
	expected := "Name: John Doe\nTitle: Manager"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_ConditionalFalse(t *testing.T) {
	template := "Name: {{name}}{{#if title}}\nTitle: {{title}}{{/if}}"
	data := map[string]string{
		"name": "John Doe",
		// title is missing
	}
	
	result := Render(template, data)
	expected := "Name: John Doe"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_ConditionalEmpty(t *testing.T) {
	template := "Name: {{name}}{{#if title}}\nTitle: {{title}}{{/if}}"
	data := map[string]string{
		"name":  "John Doe",
		"title": "", // Empty value
	}
	
	result := Render(template, data)
	expected := "Name: John Doe"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_HTMLTemplate(t *testing.T) {
	template := `<div><strong>{{name}}</strong><br>{{title}}</div>`
	data := map[string]string{
		"name":  "Jane Doe",
		"title": "Senior Engineer",
	}
	
	result := Render(template, data)
	expected := `<div><strong>Jane Doe</strong><br>Senior Engineer</div>`
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_ConditionalWithHTML(t *testing.T) {
	template := `<div>
{{#if logo}}
<img src="{{logo}}" alt="Logo">
{{/if}}
<strong>{{name}}</strong>
</div>`
	
	data := map[string]string{
		"name": "John Doe",
		"logo": "https://example.com/logo.png",
	}
	
	result := Render(template, data)
	
	if !containsSubstring(result, `<img src="https://example.com/logo.png" alt="Logo">`) {
		t.Errorf("Expected result to contain image tag, got %q", result)
	}
	if !containsSubstring(result, "<strong>John Doe</strong>") {
		t.Errorf("Expected result to contain name, got %q", result)
	}
}

func TestSanitizeHTML_BasicEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Escape less than",
			input:    "<script>",
			expected: "&lt;script&gt;",
		},
		{
			name:     "Escape greater than",
			input:    "> test",
			expected: "&gt; test",
		},
		{
			name:     "Escape ampersand",
			input:    "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
		{
			name:     "Escape quotes",
			input:    `"Hello"`,
			expected: "&#34;Hello&#34;",
		},
		{
			name:     "XSS attempt",
			input:    `<script>alert('xss')</script>`,
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRenderSafe_EscapesUserData(t *testing.T) {
	template := "<div>{{userInput}}</div>"
	data := map[string]string{
		"userInput": "<script>alert('xss')</script>",
	}
	
	result := RenderSafe(template, data)
	expected := "<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRenderSafe_PreservesTemplateTags(t *testing.T) {
	template := "<strong>{{name}}</strong>"
	data := map[string]string{
		"name": "John & Jane",
	}
	
	result := RenderSafe(template, data)
	expected := "<strong>John &amp; Jane</strong>"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRenderSafe_WithConditionals(t *testing.T) {
	template := `{{#if title}}<span>{{title}}</span>{{/if}}`
	data := map[string]string{
		"title": "<b>Manager</b>",
	}
	
	result := RenderSafe(template, data)
	expected := "<span>&lt;b&gt;Manager&lt;/b&gt;</span>"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRender_MultipleConditionals(t *testing.T) {
	template := `{{#if name}}Name: {{name}}{{/if}}
{{#if title}}Title: {{title}}{{/if}}
{{#if phone}}Phone: {{phone}}{{/if}}`
	
	data := map[string]string{
		"name":  "John Doe",
		"phone": "555-1234",
		// title is missing
	}
	
	result := Render(template, data)
	
	if !containsSubstring(result, "Name: John Doe") {
		t.Errorf("Expected name to be present, got %q", result)
	}
	if !containsSubstring(result, "Phone: 555-1234") {
		t.Errorf("Expected phone to be present, got %q", result)
	}
	if containsSubstring(result, "Title:") {
		t.Errorf("Expected title to be absent, got %q", result)
	}
}

func TestRender_NestedHTML(t *testing.T) {
	template := `<div style="font-family:Arial">
<strong>{{name}}</strong><br>
{{#if title}}{{title}}<br>{{/if}}
{{#if company}}{{company}}{{/if}}
</div>`
	
	data := map[string]string{
		"name":    "Jane Doe",
		"title":   "Senior Developer",
		"company": "Tech Corp",
	}
	
	result := Render(template, data)
	
	if !containsSubstring(result, "<strong>Jane Doe</strong>") {
		t.Errorf("Expected name in result, got %q", result)
	}
	if !containsSubstring(result, "Senior Developer") {
		t.Errorf("Expected title in result, got %q", result)
	}
	if !containsSubstring(result, "Tech Corp") {
		t.Errorf("Expected company in result, got %q", result)
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
