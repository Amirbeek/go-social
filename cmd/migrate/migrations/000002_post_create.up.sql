CREATE TABLE IF NOT EXISTS posts (
                                     id bigserial PRIMARY KEY,
                                     title text NOT NULL,
                                     user_id bigint NOT NULL,
                                     content text NOT NULL,
                                     tags text[],
                                     created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );
