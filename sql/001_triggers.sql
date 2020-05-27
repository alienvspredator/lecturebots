-- Write your migrate up statements here
CREATE TABLE triggers (
	word TEXT NOT NULL,
	bot TEXT NOT NULL,
PRIMARY KEY (word, bot)
);

---- create above / drop below ----
DROP TABLE triggers;
