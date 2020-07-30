package sqlstore

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"net/http"
)

var (
	DEALER_SEARCH_TYPE_NAME = []string{"Name"}
)

type SqlDealerStore struct {
	SqlStore
	metrics einterfaces.MetricsInterface

	// dealersQuery is a starting point for all queries that return one or more Dealers.
	dealersQuery sq.SelectBuilder
}

func (ds SqlDealerStore) ClearCaches() {}

func newSqlDealerStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.DealerStore {
	ds := &SqlDealerStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	// note: we are providing field names explicitly here to maintain order of columns (needed when using raw queries)
	ds.dealersQuery = ds.getQueryBuilder().
		Select("d.Id", "d.CreateAt", "d.UpdateAt", "d.DeleteAt", "d.Name", "d.PhoneNumber", "d.Address",
			"d.City", "d.Province", "d.Country", "d.PostalCode", "d.Brands").
		From("Dealer d")

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Dealer{}, "Dealer").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(64)
		table.ColMap("PhoneNumber").SetMaxSize(32)
		table.ColMap("Address").SetMaxSize(255)
		table.ColMap("City").SetMaxSize(64)
		table.ColMap("Province").SetMaxSize(16)
		table.ColMap("Country").SetMaxSize(5)
		table.ColMap("PostalCode").SetMaxSize(10)
		table.ColMap("Brands").SetMaxSize(255)
	}

	return ds
}

func (ds SqlDealerStore) createIndexesIfNotExists() {
	ds.CreateIndexIfNotExists("idx_dealer_update_at", "Dealer", "UpdateAt")
	ds.CreateIndexIfNotExists("idx_dealer_create_at", "Dealer", "CreateAt")
	ds.CreateIndexIfNotExists("idx_dealer_delete_at", "Dealer", "DeleteAt")
}

// Get fetches the given dealer in the database.
func (ds SqlDealerStore) Get(id string) (*model.Dealer, *model.AppError) {
	failure := func(err error, id string, statusCode int) *model.AppError {
		details := "dealer_id=" + id + ", " + err.Error()
		return model.NewAppError("SqlDealerStore.Get", id, nil, details, statusCode)
	}

	query := ds.dealersQuery.Where("Id = ?", id)
	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, failure(err, "store.sql_dealer.get.app_error", http.StatusInternalServerError)
	}
	row := ds.GetReplica().Db.QueryRow(queryString, args...)

	var dealer model.Dealer
	err = row.Scan(&dealer.Id, &dealer.CreateAt, &dealer.UpdateAt, &dealer.DeleteAt, &dealer.Name,
		&dealer.PhoneNumber, &dealer.Address, &dealer.City, &dealer.Province, &dealer.Country,
		&dealer.PostalCode, &dealer.Brands)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, failure(err, store.MISSING_ACCOUNT_ERROR, http.StatusNotFound)
		}
		return nil, failure(err, "store.sql_dealer.get.app_error", http.StatusInternalServerError)

	}

	return &dealer, nil
}

// GetAll fetches from all dealers in the database.
func (ds SqlDealerStore) GetAll() ([]*model.Dealer, *model.AppError) {
	query := ds.dealersQuery.OrderBy("Name ASC")

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlDealerStore.GetAll", "store.sql_dealer.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var data []*model.Dealer
	if _, err := ds.GetReplica().Select(&data, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlDealerStore.GetAll", "store.sql_dealer.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return data, nil
}

// Save persists a new dealer to the database.
func (ds SqlDealerStore) Save(dealer *model.Dealer) (*model.Dealer, *model.AppError) {
	if len(dealer.Id) > 0 {
		return nil, model.NewAppError("SqlDealerStore.Save", "store.sql_dealer.save.existing.app_error", nil, "dealer_id="+dealer.Id, http.StatusBadRequest)
	}

	dealer.PreSave()
	if err := dealer.IsValid(); err != nil {
		return nil, err
	}

	if err := ds.GetMaster().Insert(dealer); err != nil {
		return nil, model.NewAppError("SqlDealerStore.Save", "store.sql_dealer.save.app_error", nil, "dealer_id="+dealer.Id+", "+err.Error(), http.StatusInternalServerError)
	}

	return dealer, nil
}

// Update persists an updated dealer to the database.
func (ds SqlDealerStore) Update(dealer *model.Dealer, trustedUpdateData bool) (*model.DealerUpdate, *model.AppError) {
	dealer.PreUpdate()

	if err := dealer.IsValid(); err != nil {
		return nil, err
	}

	oldDealerResult, err := ds.GetMaster().Get(model.Dealer{}, dealer.Id)
	if err != nil {
		return nil, model.NewAppError("SqlDealerStore.Update", "store.sql_dealer.update.finding.app_error", nil, "dealer_id="+dealer.Id+", "+err.Error(), http.StatusInternalServerError)
	}

	if oldDealerResult == nil {
		return nil, model.NewAppError("SqlDealerStore.Update", "store.sql_dealer.update.find.app_error", nil, "dealer_id="+dealer.Id, http.StatusBadRequest)
	}

	oldDealer := oldDealerResult.(*model.Dealer)
	dealer.CreateAt = oldDealer.CreateAt
	dealer.Name = oldDealer.Name
	dealer.PhoneNumber = oldDealer.PhoneNumber
	dealer.Address = oldDealer.Address
	dealer.City = oldDealer.City
	dealer.Province = oldDealer.Province
	dealer.Country = oldDealer.Country
	dealer.PostalCode = oldDealer.PostalCode
	dealer.Brands = oldDealer.Brands

	count, err := ds.GetMaster().Update(dealer)
	if err != nil {
		return nil, model.NewAppError("SqlDealerStore.Update", "store.sql_dealer.update.updating.app_error", nil, "dealer_id="+dealer.Id+", "+err.Error(), http.StatusInternalServerError)
	}

	if count != 1 {
		return nil, model.NewAppError("SqlDealerStore.Update", "store.sql_dealer.update.app_error", nil, fmt.Sprintf("dealer_id=%v, count=%v", dealer.Id, count), http.StatusInternalServerError)
	}

	return &model.DealerUpdate{New: dealer, Old: oldDealer}, nil
}
