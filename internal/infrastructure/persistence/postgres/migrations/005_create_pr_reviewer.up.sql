CREATE TABLE IF NOT EXISTS pr_reviewer (
    pr_id INT REFERENCES pull_request(id) ON DELETE CASCADE,
    reviewer_id INT REFERENCES "user"(id) ON DELETE RESTRICT,
    PRIMARY KEY (pr_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewer_pr ON pr_reviewer(pr_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewer_reviewer ON pr_reviewer(reviewer_id);
