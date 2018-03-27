package bolt

import (
	"encoding/json"
	"webup/syshealth"
	"webup/syshealth/jwttools"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

var (
	bucketServers = []byte("servers")
)

// GetServerRepository returns a new bolt server repository
func GetServerRepository() syshealth.ServerRepository {
	repo := serverRepository{}
	return &repo
}

type serverRepository struct {
}

func (repo *serverRepository) GetServers() ([]syshealth.Server, error) {
	db, err := GetConnection()
	if err != nil {
		return nil, errors.Wrap(err, "unable to open bolt db")
	}

	servers := []syshealth.Server{}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketServers)
		if err != nil {
			return errors.Wrap(err, "cannot create or get bucket for 'servers'")
		}

		// if the bucket doesn't exist, just return an empty slice.
		if b == nil {
			return nil
		}

		err = b.ForEach(func(k, v []byte) error {
			server := syshealth.Server{}
			err := json.Unmarshal(v, &server)
			if err != nil {
				return errors.Wrap(err, "cannot unmarshal server data for bolt db")
			}

			servers = append(servers, server)
			return nil
		})

		return err
	})

	return servers, err
}

func (repo *serverRepository) RegisterServer(server syshealth.Server) (string, error) {

	db, err := GetConnection()
	if err != nil {
		return "", errors.Wrap(err, "unable to open bolt db")
	}

	token, id, err := jwttools.GetToken(server)
	if err != nil {
		return "", err
	}

	server.ID = id

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketServers)
		if err != nil {
			return errors.Wrap(err, "cannot create or get bucket for 'servers'")
		}

		// Marshal server data into bytes.
		buf, err := json.Marshal(server)
		if err != nil {
			return errors.Wrap(err, "cannot marshal server data into json")
		}

		// Persist bytes to servers bucket.
		return b.Put([]byte(server.ID), buf)
	})

	return token, err
}

func (repo *serverRepository) RevokeServer(id string) error {

	db, err := GetConnection()
	if err != nil {
		return errors.Wrap(err, "unable to open bolt db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketServers)
		if err != nil {
			return errors.Wrap(err, "cannot create or get bucket for 'servers'")
		}

		// if the bucket doesn't exist, just return an empty slice.
		if b == nil {
			return nil
		}

		// remove the server associated with this id
		return b.Delete([]byte(id))
	})

	return err
}

func (repo *serverRepository) CheckServerIsRegistered(id string) (bool, error) {

	db, err := GetConnection()
	if err != nil {
		return false, errors.Wrap(err, "unable to open bolt db")
	}

	found := false

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketServers)
		if err != nil {
			return errors.Wrap(err, "cannot create or get bucket for 'servers'")
		}

		// if the bucket doesn't exist, just return an empty slice.
		if b == nil {
			return nil
		}

		found = b.Get([]byte(id)) != nil

		return nil
	})

	return found, err
}
