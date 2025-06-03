-- +goose Up
-- +goose StatementBegin
CREATE TABLE train_test_split (
    hash_id String,
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY hash_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS train_test_split;
-- +goose StatementEnd
