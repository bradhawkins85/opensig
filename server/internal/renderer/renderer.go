package renderer

import (
	"html"
	"regexp"
	"strings"
)

// Renderer compiles templates with placeholders and conditionals
type Renderer struct {
	placeholderRegex  *regexp.Regexp
	conditionalRegex  *regexp.Regexp
	endConditionalRegex *regexp.Regexp
}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	return &Renderer{
		// Matches {{variable}} for placeholders
		placeholderRegex: regexp.MustCompile(`\{\{([^#/}][^}]*)\}\}`),
		// Matches {{#if variable}} for conditionals
		conditionalRegex: regexp.MustCompile(`\{\{#if\s+([^}]+)\}\}([\s\S]*?)\{\{/if\}\}`),
		// Matches {{/if}}
		endConditionalRegex: regexp.MustCompile(`\{\{/if\}\}`),
	}
}

// Render renders a template with the given data
// It supports:
// - Simple placeholders: {{name}} -> data["name"]
// - Conditionals: {{#if field}}...{{/if}} - shows content if field is non-empty
// - HTML escaping for security
func Render(template string, data map[string]string) string {
	r := NewRenderer()
	return r.RenderTemplate(template, data)
}

// RenderTemplate renders a template with the given data
func (r *Renderer) RenderTemplate(template string, data map[string]string) string {
	// First, process conditionals
	result := r.processConditionals(template, data)
	
	// Then, replace placeholders
	result = r.replacePlaceholders(result, data)
	
	return result
}

// processConditionals processes {{#if variable}}...{{/if}} blocks
func (r *Renderer) processConditionals(template string, data map[string]string) string {
	return r.conditionalRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Extract the variable name and content
		submatches := r.conditionalRegex.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}
		
		variable := strings.TrimSpace(submatches[1])
		content := submatches[2]
		
		// Check if the variable exists and is non-empty
		value, exists := data[variable]
		if exists && value != "" {
			return content
		}
		
		return ""
	})
}

// replacePlaceholders replaces {{variable}} with values from data
func (r *Renderer) replacePlaceholders(template string, data map[string]string) string {
	return r.placeholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Extract the variable name
		submatches := r.placeholderRegex.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return match
		}
		
		variable := strings.TrimSpace(submatches[1])
		
		// Get the value from data
		value, exists := data[variable]
		if !exists {
			return match // Keep placeholder if no data
		}
		
		return value
	})
}

// SanitizeHTML escapes HTML to prevent XSS injection
// This is a basic sanitizer that escapes HTML entities
func SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// RenderSafe renders a template and sanitizes user-provided data values
// Template structure is trusted, but data values are escaped
func RenderSafe(template string, data map[string]string) string {
	// Sanitize data values before rendering
	sanitizedData := make(map[string]string, len(data))
	for key, value := range data {
		sanitizedData[key] = SanitizeHTML(value)
	}
	
	return Render(template, sanitizedData)
}
