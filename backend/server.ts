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
        const { code } = req.body;

        if (!code) {
            return res.status(400).json({ error: 'No code provided' });
        }

        const response = await openai.chat.completions.create({
            model: "gpt-4",
            messages: [
                {
                    role: "system",
                    content: "You are an expert coding assistant. Analyze the provided code and suggest alternative approaches or pathways the developer could have taken. Consider different architectural patterns, libraries, or methodologies that might be more efficient or maintainable."
                },
                {
                    role: "user",
                    content: `Here's the current code implementation. Please suggest alternative approaches or pathways:\n\n${code}`
                }
            ],
            max_tokens: 1000
        });

        res.json({ analysis: response.choices[0].message.content });
    } catch (error) {
        console.error('Error analyzing code:', error);
        res.status(500).json({ error: 'Failed to analyze code' });
    }
});

app.listen(port, () => {
    console.log(`Server running on port ${port}`);
}); 