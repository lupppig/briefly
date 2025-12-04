package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

func (s *Service) AiGenResponse(ctx context.Context, text, source string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`
You are a professional content summarization AI.

The content below was extracted from: %s

Your tasks:
1. Read the content carefully and extract the most important points.
2. Summarize the content in **clear, concise, and coherent paragraphs**.
3. Focus on the **core ideas, key facts, and main messages** only.
4. Ignore filler words, background noise, timestamps, speaker labels, and metadata.
5. If the content is conversational (like a video or audio), summarize the key points as if explaining to someone who hasn’t seen it.
6. If the content contains multiple topics, organize them logically in the summary.
7. Your response should be **plain text only**, no markdown, no lists, no headings.

Content:
%s

Return ONLY the summary.
`, source, text)

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)
	if err != nil {
		return "", err
	}

	var summary strings.Builder
	for _, c := range resp.Candidates {
		for _, part := range c.Content.Parts {
			summary.WriteString(string(part.Text))
		}
	}

	return summary.String(), nil
}
