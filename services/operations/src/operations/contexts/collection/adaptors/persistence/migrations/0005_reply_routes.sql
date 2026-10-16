-- Retain the originating API connection with immutable command reply bytes.
ALTER TABLE cafe.command_replies ADD COLUMN reply_route text NOT NULL DEFAULT '';
