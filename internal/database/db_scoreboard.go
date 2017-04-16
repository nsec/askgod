package database

import (
	"database/sql"

	"github.com/lib/pq"

	"github.com/nsec/askgod/api"
)

// GetScoreboard generates the current scoreboard
func (db *DB) GetScoreboard(team *api.AdminTeam) ([]api.ScoreboardEntry, error) {
	// Return a list of score entries
	resp := []api.ScoreboardEntry{}

	// Query all the scores from the database
	var rows *sql.Rows
	var err error

	if team == nil {
		rows, err = db.Query("SELECT team.id, team.country, team.name, team.website, COALESCE(SUM(score.value), 0) AS points, MAX(score.submit_time) FROM score RIGHT JOIN team ON team.id=score.teamid WHERE team.name != '' AND team.country != '' GROUP BY team.id ORDER BY points DESC, team.id ASC;")
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = db.Query("SELECT team.id, team.country, team.name, team.website, COALESCE(SUM(score.value), 0) AS points, MAX(score.submit_time) FROM score RIGHT JOIN team ON team.id=score.teamid WHERE team.id=$1 AND team.name != '' AND team.country != '' GROUP BY team.id ORDER BY points DESC, team.id ASC;", team.ID)
		if err != nil {
			return nil, err
		}
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
