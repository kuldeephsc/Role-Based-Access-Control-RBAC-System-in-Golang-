package postgres

import "gorm.io/gorm/clause"

// onConflictDoNothing backs the AssignRole / AttachPermission upserts --
// assigning a role a user already has, or attaching a permission a role
// already has, is treated as a no-op rather than a duplicate-key error.
func onConflictDoNothing() clause.OnConflict {
	return clause.OnConflict{DoNothing: true}
}
