package database

import (
	"github.com/lib/pq"

	"github.com/nsec/askgod/api"
)

// GetScoreboard generates the current scoreboard
func (db *DB) GetScoreboard() ([]api.ScoreboardEntry, error) {
	// Return a list of score entries
	resp := []api.ScoreboardEntry{}

	// Query all the scores from the database
	rows, err := db.Query("SELECT team.id, team.country, team.name, team.website, COALESCE(SUM(score.value), 0) AS points, MAX(score.submit_time) AS last_submit_time FROM score RIGHT JOIN team ON team.id=score.teamid WHERE team.name != '' AND team.country != '' GROUP BY team.id ORDER BY points DESC, last_submit_time ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	for rows.Next() {
		row := api.ScoreboardEntry{}

		submitTime := pq.NullTime{}
		err := rows.Scan(&row.Team.ID, &row.Team.Country, &row.Team.Name, &row.Team.Website, &row.Value, &submitTime)
		if err != nil {
			return nil, err
		}

		row.LastSubmitTime = submitTime.Time

		resp = append(resp, row)
	}

	// Check for any error that might have happened
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
