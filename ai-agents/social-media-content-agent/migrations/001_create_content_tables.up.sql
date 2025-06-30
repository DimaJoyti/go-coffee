-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types
CREATE TYPE content_type AS ENUM (
    'post', 'story', 'reel', 'video', 'image', 'carousel', 
    'poll', 'live', 'event', 'product', 'article', 'blog'
);

CREATE TYPE content_format AS ENUM (
    'text', 'image', 'video', 'audio', 'document', 'link', 'mixed'
);

CREATE TYPE content_status AS ENUM (
    'draft', 'review', 'approved', 'scheduled', 'published', 
    'failed', 'archived', 'expired'
);

CREATE TYPE content_priority AS ENUM (
    'low', 'medium', 'high', 'urgent'
);

CREATE TYPE content_category AS ENUM (
    'marketing', 'educational', 'entertainment', 'news', 
    'promotional', 'user_generated', 'behind_scenes', 'product_showcase'
);

CREATE TYPE content_tone AS ENUM (
    'professional', 'casual', 'friendly', 'formal', 'humorous', 
    'inspirational', 'urgent', 'empathetic'
);

CREATE TYPE platform_type AS ENUM (
    'instagram', 'facebook', 'twitter', 'linkedin', 'tiktok', 
    'youtube', 'pinterest', 'snapchat', 'reddit'
);

CREATE TYPE content_sentiment AS ENUM (
    'positive', 'negative', 'neutral', 'mixed'
);

CREATE TYPE media_type AS ENUM (
    'image', 'video', 'audio', 'document', 'gif', 'thumbnail'
);

-- Brands table
CREATE TABLE brands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    industry VARCHAR(100),
    website VARCHAR(500),
    logo VARCHAR(500),
    colors JSONB,
    typography JSONB,
    voice JSONB,
    guidelines JSONB,
    contact_info JSONB,
    keywords TEXT[],
    tags TEXT[],
    languages TEXT[],
    time_zone VARCHAR(50) DEFAULT 'UTC',
    status VARCHAR(20) DEFAULT 'active',
    custom_fields JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    external_ids JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    version BIGINT DEFAULT 1
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200),
    avatar VARCHAR(500),
    role VARCHAR(50) NOT NULL,
    permissions TEXT[],
    brand_access UUID[],
    status VARCHAR(20) DEFAULT 'active',
    time_zone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    preferences JSONB,
    custom_fields JSONB DEFAULT '{}',
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    version BIGINT DEFAULT 1
);

-- Campaigns table
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    priority VARCHAR(20) DEFAULT 'medium',
    category VARCHAR(50),
    brand_id UUID NOT NULL REFERENCES brands(id),
    manager_id UUID NOT NULL REFERENCES users(id),
    platforms platform_type[],
    target_audience JSONB,
    budget JSONB,
    timeline JSONB,
    objectives JSONB,
    kpis JSONB,
    analytics JSONB,
    hashtags TEXT[],
    keywords TEXT[],
    tags TEXT[],
    tone content_tone DEFAULT 'friendly',
    language VARCHAR(10) DEFAULT 'en',
    approval_workflow JSONB,
    compliance JSONB,
    custom_fields JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    external_ids JSONB DEFAULT '{}',
    is_template BOOLEAN DEFAULT false,
    template_id UUID,
    is_archived BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    version BIGINT DEFAULT 1
);

-- Content table
CREATE TABLE content (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    body TEXT,
    type content_type NOT NULL,
    format content_format DEFAULT 'text',
    status content_status DEFAULT 'draft',
    priority content_priority DEFAULT 'medium',
    category content_category,
    brand_id UUID NOT NULL REFERENCES brands(id),
    campaign_id UUID REFERENCES campaigns(id),
    creator_id UUID NOT NULL REFERENCES users(id),
    approver_id UUID REFERENCES users(id),
    platforms platform_type[],
    hashtags TEXT[],
    mentions TEXT[],
    tags TEXT[],
    keywords TEXT[],
    tone content_tone DEFAULT 'friendly',
    language VARCHAR(10) DEFAULT 'en',
    scheduled_at TIMESTAMP WITH TIME ZONE,
    published_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    target_audience JSONB,
    analytics JSONB,
    custom_fields JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    external_ids JSONB DEFAULT '{}',
    is_template BOOLEAN DEFAULT false,
    template_id UUID,
    ai_generated BOOLEAN DEFAULT false,
    ai_prompt TEXT,
    ai_model VARCHAR(100),
    is_archived BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    version BIGINT DEFAULT 1
);

-- Media assets table
CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    type media_type NOT NULL,
    url VARCHAR(1000) NOT NULL,
    file_name VARCHAR(255),
    file_size BIGINT,
    mime_type VARCHAR(100),
    width INTEGER,
    height INTEGER,
    duration INTEGER, -- for videos/audio in seconds
    alt_text TEXT,
    caption TEXT,
    "order" INTEGER DEFAULT 0,
    ai_generated BOOLEAN DEFAULT false,
    ai_prompt TEXT,
    ai_model VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Content variations table
CREATE TABLE content_variations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    body TEXT NOT NULL,
    hashtags TEXT[],
    mentions TEXT[],
    platform platform_type,
    weight DECIMAL(3,2) DEFAULT 1.0,
    performance_score DECIMAL(5,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Posts table (published content)
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES content(id),
    platform platform_type NOT NULL,
    platform_post_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'draft',
    type VARCHAR(50),
    text TEXT,
    media_urls TEXT[],
    hashtags TEXT[],
    mentions TEXT[],
    link VARCHAR(1000),
    location JSONB,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    published_at TIMESTAMP WITH TIME ZONE,
    last_modified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    analytics JSONB,
    engagement JSONB,
    boost_settings JSONB,
    target_audience JSONB,
    ab_test_variant UUID,
    parent_post_id UUID,
    thread_position INTEGER DEFAULT 0,
    is_repost BOOLEAN DEFAULT false,
    original_post_id UUID,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    custom_fields JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    external_ids JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    version BIGINT DEFAULT 1
);

-- Campaign members table
CREATE TABLE campaign_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    permissions TEXT[],
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(campaign_id, user_id)
);

-- Social profiles table
CREATE TABLE social_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    username VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    bio TEXT,
    url VARCHAR(500),
    profile_image VARCHAR(500),
    cover_image VARCHAR(500),
    verified BOOLEAN DEFAULT false,
    followers BIGINT DEFAULT 0,
    following BIGINT DEFAULT 0,
    posts BIGINT DEFAULT 0,
    access_token TEXT,
    refresh_token TEXT,
    token_expiry TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    last_sync TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(brand_id, platform)
);

-- Create indexes for performance
CREATE INDEX idx_content_brand_id ON content(brand_id);
CREATE INDEX idx_content_campaign_id ON content(campaign_id);
CREATE INDEX idx_content_creator_id ON content(creator_id);
CREATE INDEX idx_content_status ON content(status);
CREATE INDEX idx_content_scheduled_at ON content(scheduled_at);
CREATE INDEX idx_content_created_at ON content(created_at);
CREATE INDEX idx_content_platforms ON content USING GIN(platforms);
CREATE INDEX idx_content_hashtags ON content USING GIN(hashtags);
CREATE INDEX idx_content_keywords ON content USING GIN(keywords);

CREATE INDEX idx_posts_content_id ON posts(content_id);
CREATE INDEX idx_posts_platform ON posts(platform);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_scheduled_at ON posts(scheduled_at);
CREATE INDEX idx_posts_published_at ON posts(published_at);

CREATE INDEX idx_campaigns_brand_id ON campaigns(brand_id);
CREATE INDEX idx_campaigns_manager_id ON campaigns(manager_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);

CREATE INDEX idx_media_assets_content_id ON media_assets(content_id);
CREATE INDEX idx_content_variations_content_id ON content_variations(content_id);

-- Create triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON brands
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_content_updated_at BEFORE UPDATE ON content
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_media_assets_updated_at BEFORE UPDATE ON media_assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_content_variations_updated_at BEFORE UPDATE ON content_variations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_social_profiles_updated_at BEFORE UPDATE ON social_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
