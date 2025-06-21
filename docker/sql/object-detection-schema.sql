-- Object Detection Service Database Schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom types
CREATE TYPE stream_type AS ENUM ('webcam', 'file', 'rtmp', 'http');
CREATE TYPE stream_status AS ENUM ('idle', 'active', 'processing', 'error', 'stopped');

-- Video streams table
CREATE TABLE video_streams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    source TEXT NOT NULL,
    type stream_type NOT NULL,
    status stream_status NOT NULL DEFAULT 'idle',
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_frame_at TIMESTAMP WITH TIME ZONE
);

-- Detection models table
CREATE TABLE detection_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    file_path TEXT NOT NULL,
    classes JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, version)
);

-- Detection results table
CREATE TABLE detection_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stream_id UUID NOT NULL REFERENCES video_streams(id) ON DELETE CASCADE,
    frame_id VARCHAR(255) NOT NULL,
    objects JSONB NOT NULL DEFAULT '[]',
    process_time_ms INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    frame_width INTEGER NOT NULL,
    frame_height INTEGER NOT NULL
);

-- Detected objects table (normalized from detection_results.objects)
CREATE TABLE detected_objects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    detection_result_id UUID NOT NULL REFERENCES detection_results(id) ON DELETE CASCADE,
    class VARCHAR(100) NOT NULL,
    confidence DECIMAL(5,4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    bounding_box JSONB NOT NULL,
    tracking_id UUID,
    stream_id UUID NOT NULL REFERENCES video_streams(id) ON DELETE CASCADE,
    frame_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Object tracking data table
CREATE TABLE tracking_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_class VARCHAR(100) NOT NULL,
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    trajectory JSONB NOT NULL DEFAULT '[]',
    velocity JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    stream_id UUID NOT NULL REFERENCES video_streams(id) ON DELETE CASCADE
);

-- Detection alerts table
CREATE TABLE detection_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stream_id UUID NOT NULL REFERENCES video_streams(id) ON DELETE CASCADE,
    object_id UUID REFERENCES detected_objects(id) ON DELETE SET NULL,
    alert_type VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'medium',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    acknowledged BOOLEAN NOT NULL DEFAULT FALSE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by VARCHAR(255)
);

-- Processing statistics table
CREATE TABLE processing_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stream_id UUID NOT NULL REFERENCES video_streams(id) ON DELETE CASCADE,
    total_frames BIGINT NOT NULL DEFAULT 0,
    processed_frames BIGINT NOT NULL DEFAULT 0,
    detected_objects BIGINT NOT NULL DEFAULT 0,
    average_process_time_ms INTEGER NOT NULL DEFAULT 0,
    fps DECIMAL(8,2) NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(stream_id)
);

-- Create indexes for better performance
CREATE INDEX idx_video_streams_status ON video_streams(status);
CREATE INDEX idx_video_streams_type ON video_streams(type);
CREATE INDEX idx_video_streams_created_at ON video_streams(created_at);

CREATE INDEX idx_detection_models_active ON detection_models(is_active);
CREATE INDEX idx_detection_models_type ON detection_models(type);

CREATE INDEX idx_detection_results_stream_id ON detection_results(stream_id);
CREATE INDEX idx_detection_results_timestamp ON detection_results(timestamp);
CREATE INDEX idx_detection_results_frame_id ON detection_results(frame_id);

CREATE INDEX idx_detected_objects_stream_id ON detected_objects(stream_id);
CREATE INDEX idx_detected_objects_class ON detected_objects(class);
CREATE INDEX idx_detected_objects_tracking_id ON detected_objects(tracking_id);
CREATE INDEX idx_detected_objects_timestamp ON detected_objects(timestamp);
CREATE INDEX idx_detected_objects_confidence ON detected_objects(confidence);

CREATE INDEX idx_tracking_data_stream_id ON tracking_data(stream_id);
CREATE INDEX idx_tracking_data_active ON tracking_data(is_active);
CREATE INDEX idx_tracking_data_class ON tracking_data(object_class);
CREATE INDEX idx_tracking_data_last_seen ON tracking_data(last_seen);

CREATE INDEX idx_detection_alerts_stream_id ON detection_alerts(stream_id);
CREATE INDEX idx_detection_alerts_acknowledged ON detection_alerts(acknowledged);
CREATE INDEX idx_detection_alerts_severity ON detection_alerts(severity);
CREATE INDEX idx_detection_alerts_timestamp ON detection_alerts(timestamp);

CREATE INDEX idx_processing_stats_stream_id ON processing_stats(stream_id);
CREATE INDEX idx_processing_stats_updated_at ON processing_stats(updated_at);

-- Create triggers for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_video_streams_updated_at 
    BEFORE UPDATE ON video_streams 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_detection_models_updated_at 
    BEFORE UPDATE ON detection_models 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processing_stats_updated_at 
    BEFORE UPDATE ON processing_stats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to clean up old data
CREATE OR REPLACE FUNCTION cleanup_old_detection_data(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete old detection results and related objects
    WITH deleted AS (
        DELETE FROM detection_results 
        WHERE timestamp < NOW() - INTERVAL '1 day' * retention_days
        RETURNING id
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;
    
    -- Delete old inactive tracking data
    DELETE FROM tracking_data 
    WHERE is_active = FALSE 
    AND last_seen < NOW() - INTERVAL '1 day' * (retention_days / 2);
    
    -- Delete old acknowledged alerts
    DELETE FROM detection_alerts 
    WHERE acknowledged = TRUE 
    AND acknowledged_at < NOW() - INTERVAL '1 day' * retention_days;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create function to update processing stats
CREATE OR REPLACE FUNCTION update_processing_stats(
    p_stream_id UUID,
    p_process_time_ms INTEGER,
    p_detected_objects_count INTEGER
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO processing_stats (
        stream_id, 
        total_frames, 
        processed_frames, 
        detected_objects, 
        average_process_time_ms,
        fps,
        last_processed_at
    ) VALUES (
        p_stream_id, 
        1, 
        1, 
        p_detected_objects_count, 
        p_process_time_ms,
        CASE WHEN p_process_time_ms > 0 THEN 1000.0 / p_process_time_ms ELSE 0 END,
        NOW()
    )
    ON CONFLICT (stream_id) DO UPDATE SET
        total_frames = processing_stats.total_frames + 1,
        processed_frames = processing_stats.processed_frames + 1,
        detected_objects = processing_stats.detected_objects + p_detected_objects_count,
        average_process_time_ms = (
            (processing_stats.average_process_time_ms * processing_stats.processed_frames + p_process_time_ms) / 
            (processing_stats.processed_frames + 1)
        ),
        fps = CASE 
            WHEN (processing_stats.average_process_time_ms * processing_stats.processed_frames + p_process_time_ms) / (processing_stats.processed_frames + 1) > 0 
            THEN 1000.0 / ((processing_stats.average_process_time_ms * processing_stats.processed_frames + p_process_time_ms) / (processing_stats.processed_frames + 1))
            ELSE 0 
        END,
        last_processed_at = NOW(),
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Insert default detection model (placeholder)
INSERT INTO detection_models (name, version, type, file_path, classes, is_active) VALUES
('YOLOv5s', '6.0', 'yolo', '/app/data/models/yolov5s.onnx', 
 '["person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat", "traffic light", "fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack", "umbrella", "handbag", "tie", "suitcase", "frisbee", "skis", "snowboard", "sports ball", "kite", "baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket", "bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl", "banana", "apple", "sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair", "couch", "potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse", "remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink", "refrigerator", "book", "clock", "vase", "scissors", "teddy bear", "hair drier", "toothbrush"]'::jsonb,
 true);

-- Create a view for active streams with latest stats
CREATE VIEW active_streams_with_stats AS
SELECT 
    vs.*,
    ps.total_frames,
    ps.processed_frames,
    ps.detected_objects,
    ps.average_process_time_ms,
    ps.fps,
    ps.last_processed_at
FROM video_streams vs
LEFT JOIN processing_stats ps ON vs.id = ps.stream_id
WHERE vs.status IN ('active', 'processing');

-- Create a view for recent detection results
CREATE VIEW recent_detections AS
SELECT 
    dr.*,
    vs.name as stream_name,
    vs.type as stream_type,
    COUNT(do.id) as object_count
FROM detection_results dr
JOIN video_streams vs ON dr.stream_id = vs.id
LEFT JOIN detected_objects do ON dr.id = do.detection_result_id
WHERE dr.timestamp > NOW() - INTERVAL '1 hour'
GROUP BY dr.id, vs.name, vs.type
ORDER BY dr.timestamp DESC;

COMMIT;
