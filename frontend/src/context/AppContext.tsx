/* eslint-disable react-refresh/only-export-components */
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useAuth } from "../auth";
import {
  projectsApi,
  type CreateProjectRequest,
  type Project,
} from "../api/projects";
import {
  hardwareApi,
  type CreateHardwareRequest,
  type Hardware,
  type HardwareCheckoutRequest,
  type HardwareCheckinRequest,
} from "../api/hardware";

type AppContextValue = {
  projects: Project[];
  hardware: Hardware[];
  loadingProjects: boolean;
  loadingHardware: boolean;

  refreshProjects: () => Promise<void>;
  refreshHardware: () => Promise<void>;
  refreshAll: () => Promise<void>;

  createProject: (data: CreateProjectRequest) => Promise<Project>;
  joinProject: (projectId: string) => Promise<Project>;
  leaveProject: (projectId: string) => Promise<Project>;
  deleteProject: (projectId: string) => Promise<void>;

  createHardware: (data: CreateHardwareRequest) => Promise<Hardware>;
  deleteHardware: (id: string) => Promise<void>;
  checkoutHardware: (
    hardwareId: string,
    data: Omit<HardwareCheckoutRequest, "userId">,
  ) => Promise<Hardware>;
  checkinHardware: (
    hardwareId: string,
    data: Omit<HardwareCheckinRequest, "userId">,
  ) => Promise<Hardware>;
};

const AppContext = createContext<AppContextValue | undefined>(undefined);
// central data store of entire app
export const AppProvider = ({ children }: { children: React.ReactNode }) => {
  const { user, isAuthenticated } = useAuth();
  const userId = user?.userId ?? "";

  const [projects, setProjects] = useState<Project[]>([]);
  const [hardware, setHardware] = useState<Hardware[]>([]);
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [loadingHardware, setLoadingHardware] = useState(false);

  const refreshProjects = useCallback(async () => {
    setLoadingProjects(true);
    try {
      const { data } = await projectsApi.list();
      setProjects(data);
    } catch {
      /* keep stale data */
    } finally {
      setLoadingProjects(false);
    }
  }, []);

  const refreshHardware = useCallback(async () => {
    setLoadingHardware(true);
    try {
      const { data } = await hardwareApi.list();
      setHardware(data);
    } catch {
      /* keep stale data */
    } finally {
      setLoadingHardware(false);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    await Promise.all([refreshProjects(), refreshHardware()]);
  }, [refreshProjects, refreshHardware]);

  // Auto-fetch on login
  useEffect(() => {
    if (isAuthenticated) {
      refreshAll();
    } else {
      setProjects([]);
      setHardware([]);
    }
  }, [isAuthenticated, refreshAll]);

  const createProject = useCallback(
    async (data: CreateProjectRequest) => {
      const { data: created } = await projectsApi.create({
        ...data,
        ownerUserId: userId,
      });
      setProjects((prev) => [created, ...prev]);
      return created;
    },
    [userId],
  );

  const joinProject = useCallback(
    async (projectId: string) => {
      const { data: updated } = await projectsApi.join(projectId, userId);
      setProjects((prev) =>
        prev.map((p) => (p.projectId === projectId ? updated : p)),
      );
      return updated;
    },
    [userId],
  );

  const leaveProject = useCallback(
    async (projectId: string) => {
      const { data: updated } = await projectsApi.leave(projectId, userId);
      setProjects((prev) =>
        prev.map((p) => (p.projectId === projectId ? updated : p)),
      );
      return updated;
    },
    [userId],
  );

  const deleteProject = useCallback(
    async (projectId: string) => {
      await projectsApi.delete(projectId, userId);
      setProjects((prev) => prev.filter((p) => p.projectId !== projectId));
      // Hardware availability may have changed after cascade
      await refreshHardware();
    },
    [userId, refreshHardware],
  );

  const createHardwareItem = useCallback(
    async (data: CreateHardwareRequest) => {
      const { data: created } = await hardwareApi.create(data);
      setHardware((prev) => [created, ...prev]);
      return created;
    },
    [],
  );

  const deleteHardwareItem = useCallback(async (id: string) => {
    await hardwareApi.delete(id);
    setHardware((prev) => prev.filter((h) => h._id !== id));
  }, []);

  const checkoutHardwareItem = useCallback(
    async (
      hardwareId: string,
      data: Omit<HardwareCheckoutRequest, "userId">,
    ) => {
      const { data: updated } = await hardwareApi.checkout(hardwareId, {
        ...data,
        userId,
      });
      setHardware((prev) =>
        prev.map((h) => (h._id === hardwareId ? updated : h)),
      );

      await refreshProjects();
      return updated;
    },
    [userId, refreshProjects],
  );

  const checkinHardwareItem = useCallback(
    async (
      hardwareId: string,
      data: Omit<HardwareCheckinRequest, "userId">,
    ) => {
      const { data: updated } = await hardwareApi.checkin(hardwareId, {
        ...data,
        userId,
      });
      setHardware((prev) =>
        prev.map((h) => (h._id === hardwareId ? updated : h)),
      );
      await refreshProjects();
      return updated;
    },
    [userId, refreshProjects],
  );

  const value = useMemo<AppContextValue>(
    () => ({
      projects,
      hardware,
      loadingProjects,
      loadingHardware,
      refreshProjects,
      refreshHardware,
      refreshAll,
      createProject,
      joinProject,
      leaveProject,
      deleteProject,
      createHardware: createHardwareItem,
      deleteHardware: deleteHardwareItem,
      checkoutHardware: checkoutHardwareItem,
      checkinHardware: checkinHardwareItem,
    }),
    [
      projects,
      hardware,
      loadingProjects,
      loadingHardware,
      refreshProjects,
      refreshHardware,
      refreshAll,
      createProject,
      joinProject,
      leaveProject,
      deleteProject,
      createHardwareItem,
      deleteHardwareItem,
      checkoutHardwareItem,
      checkinHardwareItem,
    ],
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};

export const useAppData = () => {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error("useAppData must be used within an AppProvider");
  return ctx;
};
