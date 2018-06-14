package bolt

import (
	"encoding/json"
	"sort"
	"webup/syshealth"
	"webup/syshealth/jwttools"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

var (
	bucketServers = []byte("servers")
)

// GetServerRepository returns a new bolt server repository
func GetServerRepository(databaseDir string) syshealth.ServerRepository {
	repo := serverRepository{
		databaseDir: databaseDir,
	}
	return &repo
}

type serverRepository struct {
	databaseDir string
}

func (repo *serverRepository) GetServers() ([]syshealth.Server, error) {
	db, err := GetConnection(repo.databaseDir)
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

	sort.Sort(serversByName(servers))

	return servers, err
}

func (repo *serverRepository) RegisterServer(server syshealth.Server, jwtSecret string) (string, error) {

	db, err := GetConnection(repo.databaseDir)
	if err != nil {
		return "", errors.Wrap(err, "unable to open bolt db")
	}

	token, id, err := jwttools.GetToken(server, jwtSecret)
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

	db, err := GetConnection(repo.databaseDir)
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

func (repo *serverRepository) GetServer(id string) (*syshealth.Server, error) {

	db, err := GetConnection(repo.databaseDir)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open bolt db")
	}

	var server *syshealth.Server

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketServers)
		if err != nil {
			return errors.Wrap(err, "cannot create or get bucket for 'servers'")
		}

		// if the bucket doesn't exist, just return an empty slice.
		if b == nil {
			return nil
		}

		data := b.Get([]byte(id))
		if data == nil {
			return nil
		}

		s := new(syshealth.Server)
		err := json.Unmarshal(data, s)
		if err != nil {
			return errors.Wrap(err, "cannot unmarshal server data for bolt db")
		}

		server = s

		return nil
	})

	return server, err
}

// servers sorting

type serversByName []syshealth.Server

func (s serversByName) Len() int {
	return len(s)
}
func (s serversByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s serversByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
