package datamanhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
)

func errorSlice(count int, err string) []*query.Result {
	errors := make([]*query.Result, count)
	for i := 0; i < count; i++ {
		errors[i] = &query.Result{Error: err}
	}
	return errors
}

func NewHTTPDatamanClient(destination string) (*HTTPDatamanClient, error) {
	return &HTTPDatamanClient{
		destination: destination,
		client:      &http.Client{},
	}, nil
}

type HTTPDatamanClient struct {
	destination string
	client      *http.Client
}

func (d *HTTPDatamanClient) DoQueries(ctx context.Context, queries []map[query.QueryType]query.QueryArgs) ([]*query.Result, error) {
	// TODO: better marshalling
	queriesMap := make([]map[query.QueryType]interface{}, len(queries))
	for i, q := range queries {
		for k, v := range q {
			queriesMap[i] = map[query.QueryType]interface{}{k: v}
		}
	}

	encQueries, err := json.Marshal(queriesMap)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(encQueries)

	// send task to node
	req, err := http.NewRequest(
		"POST",
		d.destination+"data/raw",
		bodyReader,
	)
	if err != nil {
		return nil, err
	}

	// Pass the context on
	req.WithContext(ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	results := make([]*query.Result, len(queries))
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}