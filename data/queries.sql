
# Recursive query to get the full game play states

with recursive result(state_id, parent_state_id, mode) AS (
	select state_id, parent_state_id, mode
        from ttt.states
        where state_id = 'd865b6ce-7d9e-475e-9473-0662de9a8ace'
	union all
	select s.state_id, s.parent_state_id, s.mode
		from result r, ttt.states s
		where s.parent_state_id = r.state_id
)
select * from result;
