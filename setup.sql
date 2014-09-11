-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at http://mozilla.org/MPL/2.0/

create table Users (
  Uid          serial not null unique primary key,
  Created      timestamp without time zone default (now() at time zone 'utc'),
  Email        varchar unique not null,
  Node         varchar unique not null,
  Generation   integer not null,
  ClientState  varchar not null
)
