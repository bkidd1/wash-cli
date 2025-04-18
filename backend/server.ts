import express from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import OpenAI from 'openai';

dotenv.config();

const app = express();
const port = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());

// Initialize OpenAI client
const openai = new OpenAI({
    apiKey: process.env.OPENAI_API_KEY
});

// Endpoint to analyze code
app.post('/analyze', async (req, res) => {
    try {
        const { files } = req.body;

        if (!files || !Array.isArray(files) || files.length === 0) {
            return res.status(400).json({ error: 'No files provided' });
        }

        // Analyze each file
        const analyses = await Promise.all(
            files.map(async (file) => {
                const response = await openai.chat.completions.create({
                    model: "gpt-4",
                    messages: [
                        {
                            role: "system",
                            content: `You are an expert software architect and Cursor AI assistant. Before suggesting any changes, carefully analyze the provided code and ask yourself:

1. Is the current implementation already optimal?
   - Does it follow best practices?
   - Is it performant and maintainable?
   - Are there any actual issues that need addressing?

2. Would refactoring provide meaningful benefits?
   - Would the benefits outweigh the risks of change?
   - Is the current solution actually the best approach?
   - Are there simpler alternatives that would work as well?

If the current implementation is already optimal, acknowledge this and explain why. If changes are needed, provide clear, step-by-step instructions for Cursor's AI to implement improvements. Structure your response as follows:

1. Code Analysis Summary:
   - Brief overview of current implementation
   - Assessment of whether changes are needed
   - Justification for keeping or changing the code

2. Implementation Steps (if changes are needed):
   For each improvement area, provide:
   - Specific file(s) to modify
   - Exact changes needed
   - Code snippets showing the changes
   - Explanation of why this change improves the code

3. Cursor AI Instructions (if changes are needed):
   - Clear, actionable steps for Cursor AI to follow
   - Specific commands or prompts to use
   - Order of operations for implementation

4. Verification Steps (if changes are needed):
   - How to test the changes
   - What to look for to confirm improvements
   - Potential edge cases to consider

Format your response in a way that can be directly used as instructions for Cursor's AI. Focus on practical, implementable changes that can be executed through Cursor's interface.`
                        },
                        {
                            role: "user",
                            content: `Please analyze this code from file ${file.fileName} and determine if changes are needed. If so, provide implementation instructions for Cursor AI:\n\n${file.content}`
                        }
                    ],
                    max_tokens: 2000
                });

                return {
                    fileName: file.fileName,
                    analysis: response.choices[0].message.content
                };
            })
        );

        // Combine analyses into a single response
        const combinedAnalysis = analyses.map(analysis => 
            `=== Analysis for ${analysis.fileName} ===\n\n${analysis.analysis}\n\n`
        ).join('\n');

        res.json({ analysis: combinedAnalysis });
    } catch (error) {
        console.error('Error analyzing code:', error);
        res.status(500).json({ error: 'Failed to analyze code' });
    }
});

app.listen(port, () => {
    console.log(`Server running on port ${port}`);
}); 