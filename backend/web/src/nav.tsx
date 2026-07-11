import { createContext, useContext } from "react";

export interface Nav {
  page: string;
  arg: any;
  go: (page: string, arg?: any) => void;
}

export const NavCtx = createContext<Nav>({ page: "dash", arg: null, go: () => {} });
export const useNav = () => useContext(NavCtx);
