
func (ctrl *controller) dummyFunction1(_ http.ResponseWriter, request *http.Request) (*types.GenericAPIResponse, error) {
	ctx := request.Context()
	dataStore := datastore.GetDataStore(ctx)

	scope, err := getApplicableScope(request)
	if err != nil {
		appcontext.GetLogger(ctx).WithError(err).Error("Error while getting applicable data")
		return nil, err
	}
	filters, matchCriteria, err := helpers.ExtractFilterInfoFromRequest(request, constants.ValidInputsFilterFields[:])
	if err != nil {
		appcontext.GetLogger(ctx).WithError(err).Error("Error in validating data")
		return nil, err
	}
	source, err := getApplicableSource(ctx, request, dataStore)
	if err != nil {
		appcontext.GetLogger(ctx).WithError(err).Error("Error while getting applicable data")
		return nil, err
	}
	statResolver, err := stats.StatsResolver(scope)
	if err != nil {
		appcontext.GetLogger(ctx).WithError(err).Error("Error in getting response")
		return nil, err
	}
	statsData, err := statResolver.dummyFunction2(ctx, dataStore, source.Id, filters, matchCriteria)
	if err != nil {
		appcontext.GetLogger(ctx).WithError(err).Error("Error in getting response")
		return nil, err
	}

	return &types.GenericAPIResponse{Success: true, Data: statsData, StatusCode: http.StatusOK}, nil
}


func (c *ControlStatusCount) dummyFunction2(ctx context.Context, dataStore model.DataStore, sourceId int, filters []*types.Filter, matchCriteria *string) ([]types.StatsData, error) {
	counts, err := dataStore.dummyFunction3(ctx, sourceId, &model.Filter{Filters: filters, Criteria: matchCriteria})
	if err != nil {
		return nil, err
	}
	if counts == nil {
		return []types.StatsData{}, nil
	}
	data := []types.StatsData{
		{Value: constants.QualifiedCount, Count: counts.QualifiedCount},
		{Value: constants.FailedCount, Count: counts.FailedCount},
		{Value: constants.NotApplicableCount, Count: counts.NotApplicableCount},
		{Value: constants.NotAssessedCount, Count: counts.NotAssessedCount},
		{Value: constants.AcceptedFailedCount, Count: counts.AcceptedFailedCount},
	}
	return data, nil
}

func (dataStore *DBStore) dummyFunction3(ctx context.Context, sourceId int, f *model.Filter) (*model.AssessmentCounts, error) {
	var counts *model.AssessmentCounts
	whereClause := helpers.dummyFunction4(f)

	err := dataStore.db.Model(&model.Input{}).
		Select("sum(case when dummy_table.status = 'Qualified' then 1 else 0 end) AS QualifiedCount, sum(case when dummy_table.status = 'Failed' then 1 else 0 end) AS FailedCount, sum(case when dummy_table.status = 'Not Applicable' then 1 else 0 end) AS NotApplicableCount, sum(case when dummy_table.status = 'Accepted Failed' then 1 else 0 end) AS AcceptedFailedCount, sum(case when dummy_table.status = 'Not Assessed' OR dummy_table.status is null then 1 else 0 end) AS NotAssessedCount").
		Joins("LEFT JOIN dummy_table ON mock_table.id = dummy_table.input_id").
		Where("mock_table.source_id = ? AND mock_table.is_active = 1", sourceId).
		Where(whereClause).
		Scan(&counts).Error
	return counts, errors.Wrap(err, "db error in getting status count by source data")
}

func dummyFunction4(f *model.Filter) string {
	clause := ""
	for i, filter := range f.Filters {
		filterClause := dummyFunction5(filter)
		if i == 0 {
			clause = filterClause
		} else {
			clause += fmt.Sprintf(" %s %s", filterCriteriaMap[*f.Criteria], filterClause)
		}
	}

	if clause != "" {
		clause = " " + clause
	}
	return clause
}


func dummyFunction5(filter *types.Filter) string {
	if strings.Contains(filter.Attribute, ".") {
		attributeParts := strings.Split(filter.Attribute, ".")
		parentAttribute := sanitizeAlphabetical(attributeParts[0])
		childAttribute := sanitizeAlphabetical(attributeParts[1])
		mappedAttribute := InputsFieldMap[parentAttribute]
		queryParts := make([]string, len(filter.Value))
		for i, v := range filter.Value {
			queryParts[i] = fmt.Sprintf("JSON_EXTRACT(%s, '$.%s') = '%s'", mappedAttribute, childAttribute, v)
		}
		completeQuery := "(" + strings.Join(queryParts, " OR ") + ")"
		return completeQuery
	}
	return fmt.Sprintf("%s IN ('%s')", InputsFieldMap[filter.Attribute], strings.Join(filter.Value, "', '"))
}