-- +goose Up

-- +goose StatementBegin
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'mailboxes_address_unique'
	) THEN
		ALTER TABLE mailboxes
		ADD CONSTRAINT mailboxes_address_unique UNIQUE (address);
	END IF;
END $$;
-- +goose StatementEnd

-- +goose Down

ALTER TABLE mailboxes
DROP CONSTRAINT IF EXISTS mailboxes_address_unique;