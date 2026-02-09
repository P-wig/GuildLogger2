import { api } from "./http";

export type Project = {
    _id: string; // MongoDB id 
    projectId: string; // required project id
    name: string; // required name
    description: string; // required description
    ownerUserId?: string; // optional owner user id
};

// projects api wrapper
export const projectsApi = {
    list: (ownerUserId?: string) => // fetch projects 
        api.get<Project[]>("/projects", { // http get to /api/projects
            params: ownerUserId ? { ownerUserId } : {}, // include param only if provided 
    }),
    // payload required to create a project 
    create: (project: {
        projectId: string; //required
        name: string;//required
        description: string;//required
        ownerUserId?: string;//optional
    }) => api.post<Project>("/projects", project), // http post to /api/ projects with project data
};
