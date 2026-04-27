/* eslint-disable react-refresh/only-export-components */
import React, { createContext, useContext, useMemo } from "react";

type AppContextValue = {
  // Empty for now - add your domain-specific data here
};

const AppContext = createContext<AppContextValue | undefined>(undefined);

// Central data store of entire app
export const AppProvider = ({ children }: { children: React.ReactNode }) => {
  const value = useMemo<AppContextValue>(() => ({}), []);

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};

export const useAppData = () => {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error("useAppData must be used within an AppProvider");
  return ctx;
};
