package database

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

type Knowledge struct {
	db         *clover.DB
	faissIndex *faiss.IndexFlat
	faissId    []string
}

type KnowledgeOpts struct {
	filename string
}

func New(args ...KnowledgeArgs) (*Knowledge, error) {
	opts := KnowledgeOpts{
		filename: "cloverDB",
	}
	for _, arg := range args {
		arg(&opts)
	}
	_, err := os.Stat(opts.filename)
	if errors.Is(err, fs.ErrNotExist) {
		os.MkdirAll(opts.filename, 0755)
	}
	db, err := clover.Open(opts.filename)
	if err != nil {
		return nil, err
	}
	faissIndex, err := faiss.NewIndexFlat(ollamaconnector.Ollamasettings.GetEmbeddingSize(), 1)
	if err != nil {
		return nil, err
	}
	return &Knowledge{db: db, faissIndex: faissIndex}, nil
}

func (kn *Knowledge) Close() {
	kn.db.Close()
}

type KnowledgeArgs func(*KnowledgeOpts)

func OptionWithFile(filename string) KnowledgeArgs {
	return func(kn *KnowledgeOpts) {
		kn.filename = filename
	}

}

func (kn *Knowledge) CreateIndex(collections []string) (err error) {
	if len(collections) == 0 {
		collections, err = kn.db.ListCollections()
	}
	if err != nil {
		return err
	}
	for _, collection := range collections {
		kn.db.ForEach(query.NewQuery(collection), func(doc *document.Document) bool {
			embFromDB := doc.Get("EmbeddingVec").([]interface{})
			emb := make([]float32, ollamaconnector.Ollamasettings.GetEmbeddingSize())
			if len(embFromDB) != len(emb) {
				panic(fmt.Sprintf("wrong embedding dimensions faiss: %d emb: %d", len(embFromDB), len(emb)))
			}
			for i := range embFromDB {
				emb[i] = float32(embFromDB[i].(float64))
			}
			err := kn.faissIndex.Add(emb)
			if err != nil {
				panic("failed to add document to faiss index")
			}
			kn.faissId = append(kn.faissId, doc.ObjectId())
			return true
		})
	}
	// \TODO close db
	// kn.db.Close()
	return
}
