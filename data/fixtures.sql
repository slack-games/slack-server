
-- Fixtures data for tables

INSERT INTO gms.users (user_id, team_id, name, team_domain) VALUES ('U000000000', 'T00000000', 'AI Bill', 'team-ai');
INSERT INTO gms.users (user_id, team_id, name, team_domain) VALUES ('U000000001', 'T00000001', 'Jim', 'well-a');

INSERT INTO gms.states (state, turn, mode, first_user_id, second_user_id) VALUES ('000000000', 'U000000000', 'Start', 'U000000000', 'U000000001');

INSERT INTO hng.states (word, guess, current, mode, user_id, parent_state_id)
VALUES ('make', 'f', '_ake', 'Turn', 'U000000001', '00000000-0000-0000-0000-000000000000');
