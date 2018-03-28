package bolt

import (
	"encoding/json"
	"webup/syshealth"
	"webup/syshealth/hash"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

var (
	bucketAdminUsers = []byte("admin_users")
)

// GetAdminUserRepository returns a new bolt admin user repository
func GetAdminUserRepository() syshealth.AdminUserRepository {
	repo := adminUserRepository{}
	return &repo
}

type adminUserRepository struct {
}

// adminUser is used to store admin credentials into DB
type adminUser struct {
	Username       string
	HashedPassword string
}

func (repo *adminUserRepository) IsSetup() (bool, error) {

	db, err := GetConnection()
	if err != nil {
		return false, errors.Wrap(err, "unable to open bolt db")
	}

	isSetup := false

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAdminUsers)

		if b == nil {
			return nil
		}

		count := b.Stats().KeyN
		if count == 0 {
			return nil
		}

		isSetup = true

		return nil
	})

	return isSetup, err
}

func (repo *adminUserRepository) Login(username string, password string) (bool, error) {

	db, err := GetConnection()
	if err != nil {
		return false, errors.Wrap(err, "unable to open bolt db")
	}

	user := new(adminUser)

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAdminUsers)
		if b == nil {
			return nil
		}

		data := b.Get([]byte(username))
		if data == nil {
			return nil
		}

		err := json.Unmarshal(data, user)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return false, errors.Wrap(err, "unable to fetch user from db")
	}

	// check if user was found
	if user == nil {
		return false, nil
	}

	return hash.Check(password, user.HashedPassword), nil
}

func (repo *adminUserRepository) GetUsers() ([]string, error) {

	db, err := GetConnection()
	if err != nil {
		return nil, errors.Wrap(err, "unable to open bolt db")
	}

	users := []string{}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAdminUsers)
		if b == nil {
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			users = append(users, string(k))
			return nil
		})

		return err
	})

	return users, err
}

func (repo *adminUserRepository) Create(username string, password string) error {

	db, err := GetConnection()
	if err != nil {
		return errors.Wrap(err, "unable to open bolt db")
	}

	// hash the password
	hash, err := hash.Create(password)
	if err != nil {
		return errors.Wrap(err, "unable to hash the password")
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucketIfNotExists(bucketAdminUsers)
		if err != nil {
			return errors.Wrap(err, "cannot create or get bucket for 'admin_users'")
		}

		user := adminUser{
			Username:       username,
			HashedPassword: hash,
		}
		buf, err := json.Marshal(user)
		if err != nil {
			return errors.Wrap(err, "cannot marshal user data into json")
		}

		return b.Put([]byte(username), buf)
	})

	return err
}

func (repo *adminUserRepository) Delete(username string) error {

	db, err := GetConnection()
	if err != nil {
		return errors.Wrap(err, "unable to open bolt db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAdminUsers)
		if b == nil {
			return nil
		}

		return b.Delete([]byte(username))
	})

	return err
}
