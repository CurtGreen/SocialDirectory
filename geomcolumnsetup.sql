SELECT AddGeometryColumn('addresses', 'geom', 4326, 'POINT', 2);

CREATE INDEX idx_organization_addresses ON addresses USING gist(geom);
