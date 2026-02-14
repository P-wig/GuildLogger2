import type { Project } from "../api/projects";
import type { Hardware } from "../api/hardware";
import styles from "./homePage.module.css";
import { ProjectCard } from "../components";
import { useState } from "react";

// ── Mock data using the real API types ──────────────────────────

const MOCK_HARDWARE: Hardware[] = [
  {
    _id: "h1",
    hardwareName: "Arduino Uno",
    capacity: 6,
    available: 2,
    assignedProjects: ["testID1"],
  },
  {
    _id: "h2",
    hardwareName: "Raspberry Pi 4",
    capacity: 10,
    available: 7,
    assignedProjects: ["testID1"],
  },
  {
    _id: "h3",
    hardwareName: "ESP32 Module",
    capacity: 4,
    available: 0,
    assignedProjects: ["testID2"],
  },
  {
    _id: "h4",
    hardwareName: "FPGA Board",
    capacity: 12,
    available: 5,
    assignedProjects: ["testID2"],
  },
  {
    _id: "h5",
    hardwareName: "Oscilloscope",
    capacity: 3,
    available: 1,
    assignedProjects: ["testID3"],
  },
];

const MOCK_PROJECTS: Project[] = [
  {
    _id: "testID1",
    projectId: "proj-alpha",
    projectName: "Project Alpha",
    description: "Fake project description 1",
    ownerUserId: "testUser1",
    assignedUsers: ["testUser1", "testUser2", "testUser3"],
    assignedHardware: [
      { hardwareId: "h1", amount: 4 },
      { hardwareId: "h2", amount: 3 },
    ],
  },
  {
    _id: "testID2",
    projectId: "proj-beta",
    projectName: "Project Beta",
    description: "Fake project description 2",
    ownerUserId: "testUser4",
    assignedUsers: ["testUser4", "testUser5"],
    assignedHardware: [
      { hardwareId: "h3", amount: 4 },
      { hardwareId: "h4", amount: 7 },
    ],
  },
  {
    _id: "testID3",
    projectId: "proj-gamma",
    projectName: "Project Gamma",
    description: "Fake project description 3",
    ownerUserId: "testUser6",
    assignedUsers: ["testUser6", "testUser7", "testUser8", "testUser9"],
    assignedHardware: [{ hardwareId: "h5", amount: 2 }],
  },
];

/** Given a project, return the subset of mock hardware assigned to it. */
function getHardwareForProject(project: Project): Hardware[] {
  const hwIds = project.assignedHardware.map((ah) => ah.hardwareId);
  return MOCK_HARDWARE.filter((hw) => hwIds.includes(hw._id));
}

export const Home = () => {
  const [projects, setProjects] = useState<Project[]>(MOCK_PROJECTS);

  const handleButtonClick = (projectId: string) => {
    setProjects((prev) =>
      prev.map((project) =>
        project._id === projectId
          ? {
              ...project,
              assignedUsers: project.assignedUsers.includes("testUser1")
                ? project.assignedUsers.filter((u) => u !== "testUser1")
                : [...project.assignedUsers, "testUser1"],
            }
          : project,
      ),
    );
  };

  return (
    <div className={styles.root}>
      <div className={styles.grid}>
        {projects.map((project) => (
          <ProjectCard
            key={project._id}
            project={project}
            hardware={getHardwareForProject(project)}
            buttonLabel={
              project.assignedUsers.includes("testUser1")
                ? "Leave project"
                : "Join project"
            }
            onButtonClick={() => handleButtonClick(project._id)}
          />
        ))}
      </div>
    </div>
  );
};
