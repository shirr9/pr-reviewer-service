CREATE TABLE IF NOT EXISTS pr_reviewer (
    pr_id  VARCHAR(255),
    reviewer_id VARCHAR(255),
    PRIMARY KEY (pr_id, reviewer_id),
    FOREIGN KEY (pr_id) REFERENCES pull_request(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES "user"(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewer_pr ON pr_reviewer(pr_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewer_reviewer ON pr_reviewer(reviewer_id);
