package database

import (
	"github.com/nsec/askgod/api"
)

// GetScoreboard generates the current scoreboard
func (db *DB) GetScoreboard() ([]api.ScoreboardEntry, error) {
	// Return a list of score entries
	resp := []api.ScoreboardEntry{}

	// Query all the scores from the database
	rows, err := db.Query("SELECT team.id, team.country, team.name, team.website, SUM(score.value) AS points, MAX(score.submit_time) FROM score LEFT JOIN team ON team.id=score.teamid GROUP BY team.id ORDER BY points DESC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	for rows.Next() {
		row := api.ScoreboardEntry{}

		err := rows.Scan(&row.Team.ID, &row.Team.Country, &row.Team.Name, &row.Team.Website, &row.Value, &row.LastSubmitTime)
		if err != nil {
			return nil, err
		}

		resp = append(resp, row)
	}

	// Check for any error that might have happened
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
