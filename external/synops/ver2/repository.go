package main

import "intelligentBI/entity"

// UserActivityRepository persists a parsed user activity event.
//
// This worker depends only on this interface, never on a concrete repository
// package — whatever satisfies it (SQL Server, or anything else later) can be
// swapped in without this file changing.
type UserActivityRepository interface {
	PersistUserActivity(event entity.UserActivityEvent) error
}
