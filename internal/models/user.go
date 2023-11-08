package models

// User represents a user entity.
// swagger:model
type User struct {
    // ID of the user.
    // required: true
    // example: 1
    ID int `json:"id"`

    // Username of the user.
    // required: true
    // example: johndoe
    Username string `json:"username"`

    // Password of the user.
    // required: true
    // The password field will not be included in the JSON response for security reasons.
    Password string `json:"-"`
}