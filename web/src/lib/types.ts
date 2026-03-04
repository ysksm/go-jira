// --- Config ---

export interface Settings {
	jira?: JiraConfig;
	jiraEndpoints?: JiraEndpoint[];
	activeEndpoint?: string;
	projects: ProjectConfig[];
	database: DatabaseConfig;
	embeddings?: EmbeddingsConfig;
	log?: LogConfig;
	sync?: SyncSettings;
	debugMode?: boolean;
}

export interface JiraConfig {
	endpoint: string;
	username: string;
	apiKey: string;
}

export interface JiraEndpoint {
	name: string;
	displayName?: string;
	endpoint: string;
	username: string;
	apiKey: string;
}

export interface ProjectConfig {
	id: string;
	key: string;
	name: string;
	syncEnabled: boolean;
	lastSynced?: string;
	endpoint?: string;
	syncCheckpoint?: SyncCheckpoint;
	snapshotCheckpoint?: SnapshotCheckpoint;
}

export interface DatabaseConfig {
	path?: string;
	databaseDir: string;
}

export interface EmbeddingsConfig {
	provider: string;
	modelName?: string;
	endpoint?: string;
	autoGenerate: boolean;
}

export interface LogConfig {
	fileEnabled: boolean;
	fileDir?: string;
	level: string;
	maxFiles: number;
}

export interface SyncSettings {
	incrementalSyncEnabled: boolean;
	incrementalSyncMarginMinutes: number;
}

export interface SyncCheckpoint {
	lastIssueUpdatedAt: string;
	lastIssueKey: string;
	itemsProcessed: number;
	totalItems: number;
}

export interface SnapshotCheckpoint {
	lastIssueId: string;
	lastIssueKey: string;
	issuesProcessed: number;
	totalIssues: number;
	snapshotsGenerated: number;
}

// --- Project ---

export interface Project {
	id: string;
	key: string;
	name: string;
	description?: string;
	lead?: string;
	url?: string;
	projectTypeKey?: string;
	style?: string;
	isPrivate?: boolean;
	endpoint?: string;
}

// --- Issue ---

export interface Issue {
	id: string;
	key: string;
	projectId: string;
	summary: string;
	description?: string;
	status?: string;
	statusCategory?: string;
	priority?: string;
	issueType?: string;
	assignee?: string;
	reporter?: string;
	created?: string;
	updated?: string;
	resolved?: string;
	dueDate?: string;
	labels?: string[];
	components?: string[];
	fixVersions?: string[];
	sprint?: string;
	storyPoints?: number;
	team?: string;
	parentKey?: string;
	subtaskKeys?: string[];
	linkedIssueKeys?: string[];
	commentCount?: number;
	attachmentCount?: number;
	isDeleted?: boolean;
	syncedAt?: string;
}

// --- Change History ---

export interface ChangeHistoryItem {
	id: string;
	issueId: string;
	issueKey: string;
	author: string;
	changedAt: string;
	field: string;
	fieldType: string;
	fromValue?: string;
	toValue?: string;
	fromString?: string;
	toString?: string;
}

// --- Sync ---

export interface SyncResult {
	projectKey: string;
	issueCount: number;
	metadataUpdated: boolean;
	duration: number;
	success: boolean;
	error?: string;
}

export interface SyncProgress {
	projectKey: string;
	phase: string;
	current: number;
	total: number;
	message: string;
}

// --- Metadata ---

export interface ProjectMetadata {
	statuses: Status[];
	priorities: Priority[];
	issueTypes: IssueType[];
	labels: Label[];
	components: Component[];
	fixVersions: FixVersion[];
}

export interface Status {
	id: string;
	name: string;
	category: string;
}

export interface Priority {
	id: string;
	name: string;
	iconUrl?: string;
}

export interface IssueType {
	id: string;
	name: string;
	description?: string;
	subtask: boolean;
	iconUrl?: string;
}

export interface Label {
	name: string;
}

export interface Component {
	id: string;
	name: string;
	description?: string;
}

export interface FixVersion {
	id: string;
	name: string;
	description?: string;
	released: boolean;
	releaseDate?: string;
}

// --- SQL ---

export interface SqlTable {
	name: string;
	columns: SqlColumn[];
}

export interface SqlColumn {
	name: string;
	type: string;
	nullable: boolean;
}
