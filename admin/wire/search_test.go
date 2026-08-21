package wire

import "testing"

func intPtr(v int) *int { return &v }

func TestSearchApplicationsRequestValidate(t *testing.T) {
	tests := []struct {
		req     SearchApplicationsRequest
		name    string
		wantErr bool
	}{
		{
			name: "empty request is valid",
			req:  SearchApplicationsRequest{},
		},
		{
			name: "fully populated valid request",
			req: SearchApplicationsRequest{
				Query:  "acme",
				Facets: map[string][]string{"status": {"active", "inactive"}},
				Dates:  map[string]DateRange{"created_at": {}},
				Sort:   &SortSpec{Field: "created_at", Order: "desc"},
				Page:   &PageRequest{Number: intPtr(2), Size: intPtr(50)},
			},
		},
		{
			name:    "unknown facet field",
			req:     SearchApplicationsRequest{Facets: map[string][]string{"tier": {"pro"}}},
			wantErr: true,
		},
		{
			name:    "unknown status value",
			req:     SearchApplicationsRequest{Facets: map[string][]string{"status": {"suspended"}}},
			wantErr: true,
		},
		{
			name: "all known status values",
			req:  SearchApplicationsRequest{Facets: map[string][]string{"status": {"active", "inactive"}}},
		},
		{
			name:    "unknown date field",
			req:     SearchApplicationsRequest{Dates: map[string]DateRange{"deleted_at": {}}},
			wantErr: true,
		},
		{
			name: "known date fields",
			req:  SearchApplicationsRequest{Dates: map[string]DateRange{"created_at": {}, "updated_at": {}}},
		},
		{
			name:    "invalid sort field",
			req:     SearchApplicationsRequest{Sort: &SortSpec{Field: "name", Order: "asc"}},
			wantErr: true,
		},
		{
			name:    "invalid sort order",
			req:     SearchApplicationsRequest{Sort: &SortSpec{Field: "created_at", Order: "sideways"}},
			wantErr: true,
		},
		{
			name: "sort order is case-insensitive",
			req:  SearchApplicationsRequest{Sort: &SortSpec{Field: "updated_at", Order: "DESC"}},
		},
		{
			name:    "empty sort order rejected",
			req:     SearchApplicationsRequest{Sort: &SortSpec{Field: "created_at"}},
			wantErr: true,
		},
		{
			name:    "size above max",
			req:     SearchApplicationsRequest{Page: &PageRequest{Size: intPtr(101)}},
			wantErr: true,
		},
		{
			name: "size at max",
			req:  SearchApplicationsRequest{Page: &PageRequest{Size: intPtr(100)}},
		},
		{
			name:    "size of zero rejected when provided",
			req:     SearchApplicationsRequest{Page: &PageRequest{Size: intPtr(0)}},
			wantErr: true,
		},
		{
			name:    "page number below one",
			req:     SearchApplicationsRequest{Page: &PageRequest{Number: intPtr(0)}},
			wantErr: true,
		},
		{
			name: "page number of one",
			req:  SearchApplicationsRequest{Page: &PageRequest{Number: intPtr(1)}},
		},
		{
			name: "page object with only number uses default size",
			req:  SearchApplicationsRequest{Page: &PageRequest{Number: intPtr(3)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestSearchUsersRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     SearchUsersRequest
		wantErr bool
	}{
		{name: "empty is valid", req: SearchUsersRequest{}},
		{name: "known status facet", req: SearchUsersRequest{Facets: map[string][]string{"status": {"active", "inactive"}}}},
		{
			name:    "unknown status value",
			req:     SearchUsersRequest{Facets: map[string][]string{"status": {"banned"}}},
			wantErr: true,
		},
		{
			name:    "unknown facet field",
			req:     SearchUsersRequest{Facets: map[string][]string{"role": {"admin"}}},
			wantErr: true,
		},
		{
			name:    "unknown date field",
			req:     SearchUsersRequest{Dates: map[string]DateRange{"last_seen_at": {}}},
			wantErr: true,
		},
		{
			name:    "invalid sort field",
			req:     SearchUsersRequest{Sort: &SortSpec{Field: "email", Order: "asc"}},
			wantErr: true,
		},
		{name: "valid sort", req: SearchUsersRequest{Sort: &SortSpec{Field: "created_at", Order: "ASC"}}},
		{
			name:    "size above max",
			req:     SearchUsersRequest{Page: &PageRequest{Size: intPtr(101)}},
			wantErr: true,
		},
		{
			name:    "page number below one",
			req:     SearchUsersRequest{Page: &PageRequest{Number: intPtr(0)}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
