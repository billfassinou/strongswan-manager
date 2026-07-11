// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { createContext, useContext } from "react";

export interface Nav {
  page: string;
  arg: any;
  go: (page: string, arg?: any) => void;
}

export const NavCtx = createContext<Nav>({ page: "dash", arg: null, go: () => {} });
export const useNav = () => useContext(NavCtx);
