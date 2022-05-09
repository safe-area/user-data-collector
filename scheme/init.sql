CREATE TABLE last_user_state
(
    user_id varchar PRIMARY KEY NOT NULL,
    hex     bigint              NOT NULL,
    action  int                 NOT NULL,
);
