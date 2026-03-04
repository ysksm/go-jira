import type {
	Settings,
	JiraEndpoint,
	Project,
	Issue,
	ChangeHistoryItem,
	SyncResult,
	SyncProgress,
	ProjectMetadata,
	SqlTable
} from '$lib/types';

const BASE = '/api';

async function post<T>(endpoint: string, body?: unknown): Promise<T> {
	const res = await fetch(`${BASE}/${endpoint}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: body ? JSON.stringify(body) : '{}'
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ message: res.statusText }));
		throw new Error(err.message || res.statusText);
	}
	return res.json();
}

// --- Config ---

export async function configGet(): Promise<Settings> {
	const res = await post<{ settings: Settings }>('config.get');
	return res.settings;
}

export async function configInitialize(params: {
	endpoint: string;
	username: string;
	apiKey: string;
	databasePath?: string;
}): Promise<Settings> {
	const res = await post<{ success: boolean; settings: Settings }>('config.initialize', params);
	return res.settings;
}

export async function configUpdate(params: {
	jira?: unknown;
	database?: unknown;
	embeddings?: unknown;
	log?: unknown;
	sync?: unknown;
	addEndpoint?: JiraEndpoint;
	removeEndpoint?: string;
	setActiveEndpoint?: string;
}): Promise<Settings> {
	const res = await post<{ success: boolean; settings: Settings }>('config.update', params);
	return res.settings;
}

// --- Projects ---

export async function projectsList(): Promise<Project[]> {
	const res = await post<{ projects: Project[] }>('projects.list');
	return res.projects;
}

export async function projectsInitialize(): Promise<{ projects: Project[]; newCount: number }> {
	return post('projects.initialize');
}

export async function projectsEnable(key: string): Promise<Project> {
	const res = await post<{ project: Project }>('projects.enable', { key });
	return res.project;
}

export async function projectsDisable(key: string): Promise<Project> {
	const res = await post<{ project: Project }>('projects.disable', { key });
	return res.project;
}

// --- Sync ---

export async function syncExecute(params?: {
	projectKey?: string;
	force?: boolean;
}): Promise<SyncResult[]> {
	const res = await post<{ results: SyncResult[] }>('sync.execute', params);
	return res.results;
}

export async function syncStatus(): Promise<{
	inProgress: boolean;
	progress?: SyncProgress;
}> {
	return post('sync.status');
}

export function syncProgressSSE(
	onMessage: (data: { inProgress: boolean; progress?: SyncProgress }) => void,
	onError?: (err: Event) => void
): EventSource {
	const es = new EventSource(`${BASE}/sync.progress`);
	es.onmessage = (event) => {
		const data = JSON.parse(event.data);
		onMessage(data);
		if (!data.inProgress) {
			es.close();
		}
	};
	es.onerror = (err) => {
		onError?.(err);
		es.close();
	};
	return es;
}

// --- Issues ---

export async function issuesSearch(params: {
	project: string;
	query?: string;
	status?: string;
	assignee?: string;
	priority?: string;
	issueType?: string;
	limit?: number;
	offset?: number;
}): Promise<{ issues: Issue[]; total: number }> {
	return post('issues.search', params);
}

export async function issuesGet(key: string): Promise<Issue> {
	const res = await post<{ issue: Issue }>('issues.get', { key });
	return res.issue;
}

export async function issuesHistory(key: string): Promise<ChangeHistoryItem[]> {
	const res = await post<{ history: ChangeHistoryItem[] }>('issues.history', { key });
	return res.history;
}

// --- Metadata ---

export async function metadataGet(projectKey: string): Promise<ProjectMetadata> {
	const res = await post<{ metadata: ProjectMetadata }>('metadata.get', { projectKey });
	return res.metadata;
}

// --- SQL ---

export async function sqlExecute(params: {
	projectKey?: string;
	allProjects?: boolean;
	query: string;
	limit?: number;
}): Promise<{
	columns: string[];
	rows: unknown[][];
	rowCount: number;
	executionTimeMs: number;
}> {
	return post('sql.execute', params);
}

export async function sqlGetSchema(projectKey: string): Promise<SqlTable[]> {
	const res = await post<{ tables: SqlTable[] }>('sql.get-schema', { projectKey });
	return res.tables;
}
