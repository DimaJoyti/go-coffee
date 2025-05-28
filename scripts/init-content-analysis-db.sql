-- Content Analysis System Database Schema
-- This script initializes the database schema for the content analysis system

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS content_analysis;
CREATE SCHEMA IF NOT EXISTS reddit;
CREATE SCHEMA IF NOT EXISTS rag;

-- Set search path
SET search_path TO content_analysis, reddit, rag, public;

-- Reddit Posts table
CREATE TABLE IF NOT EXISTS reddit.posts (
    id VARCHAR(50) PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    author VARCHAR(100),
    subreddit VARCHAR(100) NOT NULL,
    url TEXT,
    score INTEGER DEFAULT 0,
    upvote_ratio DECIMAL(3,2),
    num_comments INTEGER DEFAULT 0,
    created_utc TIMESTAMP WITH TIME ZONE NOT NULL,
    is_video BOOLEAN DEFAULT FALSE,
    is_self BOOLEAN DEFAULT FALSE,
    permalink TEXT,
    flair VARCHAR(200),
    nsfw BOOLEAN DEFAULT FALSE,
    spoiler BOOLEAN DEFAULT FALSE,
    locked BOOLEAN DEFAULT FALSE,
    stickied BOOLEAN DEFAULT FALSE,
    metadata JSONB,
    processed_at TIMESTAMP WITH TIME ZONE,
    classification VARCHAR(100),
    sentiment VARCHAR(50),
    topics TEXT[],
    confidence DECIMAL(3,2),
    embedding_vector DECIMAL[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Reddit Comments table
CREATE TABLE IF NOT EXISTS reddit.comments (
    id VARCHAR(50) PRIMARY KEY,
    post_id VARCHAR(50) REFERENCES reddit.posts(id) ON DELETE CASCADE,
    parent_id VARCHAR(50),
    content TEXT NOT NULL,
    author VARCHAR(100),
    score INTEGER DEFAULT 0,
    created_utc TIMESTAMP WITH TIME ZONE NOT NULL,
    is_submitter BOOLEAN DEFAULT FALSE,
    depth INTEGER DEFAULT 0,
    permalink TEXT,
    metadata JSONB,
    processed_at TIMESTAMP WITH TIME ZONE,
    classification VARCHAR(100),
    sentiment VARCHAR(50),
    topics TEXT[],
    confidence DECIMAL(3,2),
    embedding_vector DECIMAL[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Content Classifications table
CREATE TABLE IF NOT EXISTS content_analysis.classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id VARCHAR(50) NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('post', 'comment')),
    category VARCHAR(100) NOT NULL,
    subcategory VARCHAR(100),
    tags TEXT[],
    confidence DECIMAL(3,2) NOT NULL,
    model_used VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Sentiment Analysis table
CREATE TABLE IF NOT EXISTS content_analysis.sentiment_analysis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id VARCHAR(50) NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('post', 'comment')),
    label VARCHAR(20) NOT NULL CHECK (label IN ('positive', 'negative', 'neutral')),
    score DECIMAL(3,2) NOT NULL,
    magnitude DECIMAL(3,2),
    subjectivity DECIMAL(3,2),
    model_used VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Topic Analysis table
CREATE TABLE IF NOT EXISTS content_analysis.topic_analysis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id VARCHAR(50) NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('post', 'comment')),
    topic VARCHAR(200) NOT NULL,
    keywords TEXT[],
    probability DECIMAL(3,2) NOT NULL,
    relevance DECIMAL(3,2),
    model_used VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trend Analysis table
CREATE TABLE IF NOT EXISTS content_analysis.trend_analysis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timeframe VARCHAR(20) NOT NULL,
    subreddit VARCHAR(100),
    category VARCHAR(100),
    trend_type VARCHAR(50),
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    avg_score DECIMAL(10,2),
    avg_comments DECIMAL(10,2),
    engagement_rate DECIMAL(5,4),
    growth_rate DECIMAL(5,4),
    velocity_score DECIMAL(5,4),
    trending_keywords JSONB,
    sentiment_trend JSONB,
    topics JSONB,
    predictions JSONB,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- RAG Documents table
CREATE TABLE IF NOT EXISTS rag.documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content TEXT NOT NULL,
    title VARCHAR(500),
    source VARCHAR(100) NOT NULL,
    source_id VARCHAR(100),
    metadata JSONB,
    embedding_vector DECIMAL[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    indexed_at TIMESTAMP WITH TIME ZONE,
    version INTEGER DEFAULT 1,
    tags TEXT[],
    category VARCHAR(100),
    language VARCHAR(10) DEFAULT 'en',
    word_count INTEGER,
    quality DECIMAL(3,2)
);

-- RAG Document Chunks table
CREATE TABLE IF NOT EXISTS rag.document_chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID REFERENCES rag.documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    start_offset INTEGER NOT NULL,
    end_offset INTEGER NOT NULL,
    embedding_vector DECIMAL[],
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- RAG Queries table
CREATE TABLE IF NOT EXISTS rag.queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    text TEXT NOT NULL,
    user_id VARCHAR(100),
    context TEXT,
    filters JSONB,
    embedding_vector DECIMAL[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    language VARCHAR(10) DEFAULT 'en',
    intent VARCHAR(100),
    entities JSONB
);

-- RAG Responses table
CREATE TABLE IF NOT EXISTS rag.responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID REFERENCES rag.queries(id) ON DELETE CASCADE,
    generated_text TEXT NOT NULL,
    sources JSONB,
    confidence DECIMAL(3,2),
    model_used VARCHAR(100),
    processing_time_ms INTEGER,
    tokens_used INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    citations JSONB,
    follow_up_queries TEXT[]
);

-- Processing Jobs table
CREATE TABLE IF NOT EXISTS content_analysis.processing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    content_ids TEXT[],
    filter_config JSONB,
    job_config JSONB,
    progress DECIMAL(5,2) DEFAULT 0,
    results JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

-- Feedback table
CREATE TABLE IF NOT EXISTS rag.feedback (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_id UUID REFERENCES rag.queries(id) ON DELETE CASCADE,
    response_id UUID REFERENCES rag.responses(id) ON DELETE CASCADE,
    user_id VARCHAR(100),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    helpful BOOLEAN,
    accurate BOOLEAN,
    complete BOOLEAN,
    relevant BOOLEAN,
    comments TEXT,
    improvements TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB
);

-- Analytics table
CREATE TABLE IF NOT EXISTS content_analysis.analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timeframe VARCHAR(20) NOT NULL,
    query_count INTEGER DEFAULT 0,
    avg_response_time_ms INTEGER,
    avg_confidence DECIMAL(3,2),
    avg_rating DECIMAL(3,2),
    top_queries JSONB,
    top_sources JSONB,
    error_rate DECIMAL(5,4),
    cache_hit_rate DECIMAL(5,4),
    model_usage JSONB,
    user_engagement JSONB,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB
);

-- Create indexes for performance

-- Reddit posts indexes
CREATE INDEX IF NOT EXISTS idx_posts_subreddit ON reddit.posts(subreddit);
CREATE INDEX IF NOT EXISTS idx_posts_created_utc ON reddit.posts(created_utc);
CREATE INDEX IF NOT EXISTS idx_posts_score ON reddit.posts(score);
CREATE INDEX IF NOT EXISTS idx_posts_classification ON reddit.posts(classification);
CREATE INDEX IF NOT EXISTS idx_posts_sentiment ON reddit.posts(sentiment);
CREATE INDEX IF NOT EXISTS idx_posts_processed_at ON reddit.posts(processed_at);
CREATE INDEX IF NOT EXISTS idx_posts_title_gin ON reddit.posts USING gin(to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_posts_content_gin ON reddit.posts USING gin(to_tsvector('english', content));

-- Reddit comments indexes
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON reddit.comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_utc ON reddit.comments(created_utc);
CREATE INDEX IF NOT EXISTS idx_comments_score ON reddit.comments(score);
CREATE INDEX IF NOT EXISTS idx_comments_classification ON reddit.comments(classification);
CREATE INDEX IF NOT EXISTS idx_comments_sentiment ON reddit.comments(sentiment);
CREATE INDEX IF NOT EXISTS idx_comments_content_gin ON reddit.comments USING gin(to_tsvector('english', content));

-- Content analysis indexes
CREATE INDEX IF NOT EXISTS idx_classifications_content_id ON content_analysis.classifications(content_id);
CREATE INDEX IF NOT EXISTS idx_classifications_category ON content_analysis.classifications(category);
CREATE INDEX IF NOT EXISTS idx_classifications_processed_at ON content_analysis.classifications(processed_at);

CREATE INDEX IF NOT EXISTS idx_sentiment_content_id ON content_analysis.sentiment_analysis(content_id);
CREATE INDEX IF NOT EXISTS idx_sentiment_label ON content_analysis.sentiment_analysis(label);
CREATE INDEX IF NOT EXISTS idx_sentiment_processed_at ON content_analysis.sentiment_analysis(processed_at);

CREATE INDEX IF NOT EXISTS idx_topics_content_id ON content_analysis.topic_analysis(content_id);
CREATE INDEX IF NOT EXISTS idx_topics_topic ON content_analysis.topic_analysis(topic);
CREATE INDEX IF NOT EXISTS idx_topics_processed_at ON content_analysis.topic_analysis(processed_at);

-- RAG indexes
CREATE INDEX IF NOT EXISTS idx_documents_source ON rag.documents(source);
CREATE INDEX IF NOT EXISTS idx_documents_category ON rag.documents(category);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON rag.documents(created_at);
CREATE INDEX IF NOT EXISTS idx_documents_content_gin ON rag.documents USING gin(to_tsvector('english', content));

CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON rag.document_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_chunks_chunk_index ON rag.document_chunks(chunk_index);

CREATE INDEX IF NOT EXISTS idx_queries_user_id ON rag.queries(user_id);
CREATE INDEX IF NOT EXISTS idx_queries_created_at ON rag.queries(created_at);

CREATE INDEX IF NOT EXISTS idx_responses_query_id ON rag.responses(query_id);
CREATE INDEX IF NOT EXISTS idx_responses_created_at ON rag.responses(created_at);

-- Create triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON reddit.posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON reddit.comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON rag.documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries

-- Content analysis summary view
CREATE OR REPLACE VIEW content_analysis.content_summary AS
SELECT 
    p.id,
    p.title,
    p.subreddit,
    p.author,
    p.score,
    p.created_utc,
    c.category,
    c.confidence as classification_confidence,
    s.label as sentiment,
    s.score as sentiment_score,
    array_agg(DISTINCT t.topic) as topics
FROM reddit.posts p
LEFT JOIN content_analysis.classifications c ON p.id = c.content_id AND c.content_type = 'post'
LEFT JOIN content_analysis.sentiment_analysis s ON p.id = s.content_id AND s.content_type = 'post'
LEFT JOIN content_analysis.topic_analysis t ON p.id = t.content_id AND t.content_type = 'post'
GROUP BY p.id, p.title, p.subreddit, p.author, p.score, p.created_utc, c.category, c.confidence, s.label, s.score;

-- Trending content view
CREATE OR REPLACE VIEW content_analysis.trending_content AS
SELECT 
    p.id,
    p.title,
    p.subreddit,
    p.score,
    p.num_comments,
    p.created_utc,
    (p.score + p.num_comments * 2) / EXTRACT(EPOCH FROM (NOW() - p.created_utc)) / 3600 as trend_score,
    c.category,
    s.label as sentiment
FROM reddit.posts p
LEFT JOIN content_analysis.classifications c ON p.id = c.content_id AND c.content_type = 'post'
LEFT JOIN content_analysis.sentiment_analysis s ON p.id = s.content_id AND s.content_type = 'post'
WHERE p.created_utc > NOW() - INTERVAL '24 hours'
ORDER BY trend_score DESC;

-- Grant permissions
GRANT USAGE ON SCHEMA content_analysis TO postgres;
GRANT USAGE ON SCHEMA reddit TO postgres;
GRANT USAGE ON SCHEMA rag TO postgres;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA content_analysis TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA reddit TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA rag TO postgres;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA content_analysis TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA reddit TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA rag TO postgres;

-- Insert sample data for testing
INSERT INTO reddit.posts (id, title, content, author, subreddit, score, created_utc) VALUES
('sample_1', 'Introduction to Machine Learning', 'This is a comprehensive guide to ML basics...', 'ml_expert', 'MachineLearning', 150, NOW() - INTERVAL '2 hours'),
('sample_2', 'Best Coffee Brewing Methods', 'Exploring different ways to brew the perfect cup...', 'coffee_lover', 'Coffee', 89, NOW() - INTERVAL '1 hour'),
('sample_3', 'Web3 and DeFi Trends', 'Analysis of current trends in decentralized finance...', 'crypto_analyst', 'technology', 234, NOW() - INTERVAL '30 minutes');

-- Log completion
INSERT INTO content_analysis.analytics (timeframe, query_count, generated_at, metadata) VALUES
('initialization', 0, NOW(), '{"event": "database_initialized", "version": "1.0.0"}');

COMMIT;
