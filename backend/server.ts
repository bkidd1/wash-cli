import express from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import OpenAI from 'openai';

dotenv.config();

const app = express();
const port = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json({ limit: '50mb' }));

// Initialize OpenAI client
const openai = new OpenAI({
    apiKey: process.env.OPENAI_API_KEY
});

interface FileAnalysis {
    fileName: string;
    content: string;
}

// Endpoint to analyze project structure only
app.post('/analyze-structure', async (req, res) => {
    try {
        console.log('Received project structure analysis request');
        const { projectStructure } = req.body;

        if (!projectStructure) {
            console.error('No project structure provided in request');
            return res.status(400).json({ error: 'No project structure provided' });
        }

        console.log('Project structure received:', JSON.stringify(projectStructure, null, 2));

        try {
            const structureAnalysis = await openai.chat.completions.create({
                model: "gpt-4",
                messages: [
                    {
                        role: "system",
                        content: `You are an expert software architect. Analyze the provided project structure and provide insights about:
1. Overall project organization
2. Potential improvements in file/directory structure
3. Missing or redundant components
4. Best practices and recommendations

Format your response in a clear, structured way with sections for each aspect.`
                    },
                    {
                        role: "user",
                        content: `Please analyze this project structure:\n\n${JSON.stringify(projectStructure, null, 2)}`
                    }
                ],
                max_tokens: 2000
            });

            console.log('Received analysis from OpenAI');
            res.json({ analysis: structureAnalysis.choices[0].message.content });
        } catch (openaiError: any) {
            console.error('OpenAI API Error:', {
                name: openaiError?.name,
                message: openaiError?.message,
                status: openaiError?.status,
                code: openaiError?.code,
                type: openaiError?.type
            });
            throw openaiError;
        }
    } catch (error) {
        console.error('Error in /analyze-structure endpoint:', error);
        if (error instanceof Error) {
            console.error('Error details:', {
                name: error.name,
                message: error.message,
                stack: error.stack
            });
        }
        res.status(500).json({ 
            error: 'Failed to analyze project structure',
            details: error instanceof Error ? error.message : 'Unknown error'
        });
    }
});

// Endpoint to analyze a single file
app.post('/analyze-file', async (req, res) => {
    try {
        const { fileName, content } = req.body;

        if (!fileName || !content) {
            return res.status(400).json({ error: 'File name and content are required' });
        }

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

If the current implementation is already optimal, acknowledge this and explain why. If changes are needed, provide clear, step-by-step instructions for Cursor's AI to implement improvements.`
                },
                {
                    role: "user",
                    content: `Please analyze this code from file ${fileName} and determine if changes are needed:\n\n${content}`
                }
            ],
            max_tokens: 2000
        });

        res.json({ 
            fileName,
            analysis: response.choices[0].message.content 
        });
    } catch (error) {
        console.error('Error analyzing file:', error);
        res.status(500).json({ error: 'Failed to analyze file' });
    }
});

// Endpoint to analyze chat history from screenshot
app.post('/analyze-chat', async (req, res) => {
    try {
        const { screenshot, lastAnalysisTime } = req.body;

        if (!screenshot) {
            return res.status(400).json({ error: 'Screenshot is required' });
        }

        const response = await openai.chat.completions.create({
            model: "gpt-4.1-mini",
            messages: [
                {
                    role: "system",
                    content: `You are an expert AI assistant analyzing ongoing chat history. Your task is to:
1. Identify key discussion points and decisions made
2. Extract actionable suggestions for code improvements
3. Generate clear, copy-pasteable prompts that the user can use in Cursor chat
4. Note any patterns in communication that could be improved
5. Track progress on major tasks and decisions

Format your response as follows:

KEY POINTS:
- List main discussion topics and decisions
- Highlight any blockers or challenges identified

ACTIONABLE SUGGESTIONS:
- Provide specific prompts that the user can copy into Cursor chat
- Focus on architectural improvements, best practices, and optimization opportunities
- Each suggestion should be a complete, self-contained prompt

COMMUNICATION PATTERNS:
- Note any recurring issues in how requests are framed
- Suggest better ways to phrase questions or requests
- Highlight successful communication strategies

PROGRESS TRACKING:
- Summarize what has been accomplished
- List pending items or next steps
- Note any dependencies or prerequisites

DO NOT generate code directly. Instead, provide prompts that the user can copy into Cursor chat to get the desired code.`
                },
                {
                    role: "user",
                    content: [
                        {
                            type: "text",
                            text: "Please analyze this latest segment of the ongoing chat and provide structured insights and actionable suggestions:"
                        },
                        {
                            type: "image_url",
                            image_url: {
                                url: screenshot,
                                detail: "low"  // Using low detail to save tokens since we're analyzing text
                            }
                        }
                    ]
                }
            ],
            max_tokens: 2000
        });

        res.json({ 
            analysis: response.choices[0].message.content 
        });
    } catch (error) {
        console.error('Error analyzing chat:', error);
        res.status(500).json({ error: 'Failed to analyze chat' });
    }
});

// Endpoint to summarize chat notes
app.post('/summarize-chat', async (req, res) => {
    try {
        const { notes } = req.body;

        if (!notes) {
            return res.status(400).json({ error: 'Notes content is required' });
        }

        const response = await openai.chat.completions.create({
            model: "gpt-4",
            messages: [
                {
                    role: "system",
                    content: `You are an expert AI assistant analyzing chat history summaries. Your task is to:
1. Identify the main patterns and themes in the conversation
2. Highlight recurring issues or misunderstandings
3. Note successful communication strategies
4. Provide actionable recommendations for improvement
5. Track the overall progress of the interaction

Format your response in a clear, structured way with these sections:
- Key Patterns and Themes
- Communication Strengths
- Areas for Improvement
- Actionable Recommendations
- Overall Progress`
                },
                {
                    role: "user",
                    content: `Please analyze and summarize these chat notes:\n\n${notes}`
                }
            ],
            max_tokens: 2000
        });

        res.json({ 
            summary: response.choices[0].message.content 
        });
    } catch (error) {
        console.error('Error summarizing chat:', error);
        res.status(500).json({ error: 'Failed to summarize chat' });
    }
});

// Error handling middleware
app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
    console.error('Error:', err);
    if (err.type === 'entity.too.large') {
        return res.status(413).json({ 
            error: 'Request payload too large. Please split your request into smaller chunks using the separate endpoints.' 
        });
    }
    res.status(500).json({ error: 'Internal server error' });
});

app.listen(port, () => {
    console.log(`Server running on port ${port}`);
}); 