CREATE TABLE IF NOT EXISTS pull_request (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author_id  VARCHAR(255),
    status pr_status DEFAULT 'OPEN',
    need_more_reviewers BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES "user"(id)
);

CREATE INDEX IF NOT EXISTS idx_pull_request_author ON pull_request(author_id);

