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