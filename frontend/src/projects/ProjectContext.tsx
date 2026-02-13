/* eslint-disable react-refresh/only-export-components */
import React, { createContext, useContext, useMemo, useState } from "react";


export type CurrentProject = {
    projectId: string;
    name: string;
};

type ProjectContextValue = {
    currentProject: CurrentProject | null;
    setCurrentProject: (project: CurrentProject | null) => void;
};

const ProjectContext = createContext<ProjectContextValue | undefined>(undefined);

export const ProjectProvider = ({ children }: { children: React.ReactNode }) => {
    const [currentProject, setCurrentProject] = useState<CurrentProject | null>(null);

    const value = useMemo(
        () => ({ currentProject, setCurrentProject }),
        [currentProject],
    );
    return (
        <ProjectContext.Provider value={value}>
            {children}
        </ProjectContext.Provider>
    );
};

export const useProject = () => {
    const context = useContext(ProjectContext);
    if (!context) {
        throw new Error("useProject must be used within a ProjectProvider");
    }
    return context;
}
