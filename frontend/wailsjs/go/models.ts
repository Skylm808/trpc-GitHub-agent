export namespace config {

	export class ProviderConfig {
	    name: string;
	    model: string;
	    base_url: string;
	    enabled: boolean;

	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model = source["model"];
	        this.base_url = source["base_url"];
	        this.enabled = source["enabled"];
	    }
	}
	export class AppConfig {
	    model_provider: string;
	    model_name: string;
	    ui_language: string;
	    github_base_url: string;
	    providers: ProviderConfig[];
	    default_language: string;
	    input_language: string;
	    default_direction: string;
	    default_difficulty: string;
	    default_min_stars: number;
	    default_max_stars: number;
	    default_pushed_after: string;
	    theme: string;

	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model_provider = source["model_provider"];
	        this.model_name = source["model_name"];
	        this.ui_language = source["ui_language"];
	        this.github_base_url = source["github_base_url"];
	        this.providers = this.convertValues(source["providers"], ProviderConfig);
	        this.default_language = source["default_language"];
	        this.input_language = source["input_language"];
	        this.default_direction = source["default_direction"];
	        this.default_difficulty = source["default_difficulty"];
	        this.default_min_stars = source["default_min_stars"];
	        this.default_max_stars = source["default_max_stars"];
	        this.default_pushed_after = source["default_pushed_after"];
	        this.theme = source["theme"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConnectionCheck {
	    target: string;
	    ok: boolean;
	    message: string;

	    static createFrom(source: any = {}) {
	        return new ConnectionCheck(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.ok = source["ok"];
	        this.message = source["message"];
	    }
	}

	export class ProviderConnectionRequest {
	    provider: string;
	    model: string;
	    base_url: string;
	    token: string;
	    enabled: boolean;

	    static createFrom(source: any = {}) {
	        return new ProviderConnectionRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.base_url = source["base_url"];
	        this.token = source["token"];
	        this.enabled = source["enabled"];
	    }
	}
	export class SettingsStatus {
	    github_token_configured: boolean;
	    github_base_url: string;
	    model_provider: string;
	    model_name: string;
	    openai_configured: boolean;
	    anthropic_configured: boolean;
	    deepseek_configured: boolean;
	    active_provider_ready: boolean;
	    deterministic_mode_ready: boolean;
	    providers: llm.Provider[];

	    static createFrom(source: any = {}) {
	        return new SettingsStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.github_token_configured = source["github_token_configured"];
	        this.github_base_url = source["github_base_url"];
	        this.model_provider = source["model_provider"];
	        this.model_name = source["model_name"];
	        this.openai_configured = source["openai_configured"];
	        this.anthropic_configured = source["anthropic_configured"];
	        this.deepseek_configured = source["deepseek_configured"];
	        this.active_provider_ready = source["active_provider_ready"];
	        this.deterministic_mode_ready = source["deterministic_mode_ready"];
	        this.providers = this.convertValues(source["providers"], llm.Provider);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SettingsBundle {
	    config: AppConfig;
	    status: SettingsStatus;
	    config_path: string;
	    secrets_path: string;

	    static createFrom(source: any = {}) {
	        return new SettingsBundle(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], AppConfig);
	        this.status = this.convertValues(source["status"], SettingsStatus);
	        this.config_path = source["config_path"];
	        this.secrets_path = source["secrets_path"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class SettingsUpdate {
	    config: AppConfig;
	    github_token: string;
	    provider_tokens: Record<string, string>;
	    openai_api_key: string;
	    anthropic_api_key: string;
	    deepseek_api_key: string;

	    static createFrom(source: any = {}) {
	        return new SettingsUpdate(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], AppConfig);
	        this.github_token = source["github_token"];
	        this.provider_tokens = source["provider_tokens"];
	        this.openai_api_key = source["openai_api_key"];
	        this.anthropic_api_key = source["anthropic_api_key"];
	        this.deepseek_api_key = source["deepseek_api_key"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace domain {

	export class AgentTraceStep {
	    phase: string;
	    tool: string;
	    summary: string;

	    static createFrom(source: any = {}) {
	        return new AgentTraceStep(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.phase = source["phase"];
	        this.tool = source["tool"];
	        this.summary = source["summary"];
	    }
	}
	export class Score {
	    repository_id: number;
	    activity_score: number;
	    popularity_score: number;
	    learning_value_score: number;
	    contribution_friendliness_score: number;
	    role_relevance_score: number;
	    total_score: number;
	    influence_level: string;
	    beginner_friendliness: string;
	    difficulty: string;
	    recommendation_reason: string;
	    explanation: Record<string, string>;
	    // Go type: time
	    scored_at: any;

	    static createFrom(source: any = {}) {
	        return new Score(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repository_id = source["repository_id"];
	        this.activity_score = source["activity_score"];
	        this.popularity_score = source["popularity_score"];
	        this.learning_value_score = source["learning_value_score"];
	        this.contribution_friendliness_score = source["contribution_friendliness_score"];
	        this.role_relevance_score = source["role_relevance_score"];
	        this.total_score = source["total_score"];
	        this.influence_level = source["influence_level"];
	        this.beginner_friendliness = source["beginner_friendliness"];
	        this.difficulty = source["difficulty"];
	        this.recommendation_reason = source["recommendation_reason"];
	        this.explanation = source["explanation"];
	        this.scored_at = this.convertValues(source["scored_at"], null);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Repository {
	    id: number;
	    github_id: number;
	    full_name: string;
	    owner: string;
	    name: string;
	    description: string;
	    html_url: string;
	    clone_url: string;
	    language: string;
	    topics: string[];
	    stars: number;
	    forks: number;
	    watchers: number;
	    open_issues_count: number;
	    default_branch: string;
	    archived: boolean;
	    disabled: boolean;
	    // Go type: time
	    pushed_at: any;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    // Go type: time
	    fetched_at: any;

	    static createFrom(source: any = {}) {
	        return new Repository(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.github_id = source["github_id"];
	        this.full_name = source["full_name"];
	        this.owner = source["owner"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.html_url = source["html_url"];
	        this.clone_url = source["clone_url"];
	        this.language = source["language"];
	        this.topics = source["topics"];
	        this.stars = source["stars"];
	        this.forks = source["forks"];
	        this.watchers = source["watchers"];
	        this.open_issues_count = source["open_issues_count"];
	        this.default_branch = source["default_branch"];
	        this.archived = source["archived"];
	        this.disabled = source["disabled"];
	        this.pushed_at = this.convertValues(source["pushed_at"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.fetched_at = this.convertValues(source["fetched_at"], null);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ScoredRepository {
	    repository: Repository;
	    score: Score;

	    static createFrom(source: any = {}) {
	        return new ScoredRepository(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repository = this.convertValues(source["repository"], Repository);
	        this.score = this.convertValues(source["score"], Score);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlannedQuery {
	    query: string;
	    reason: string;
	    description: string;

	    static createFrom(source: any = {}) {
	        return new PlannedQuery(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.reason = source["reason"];
	        this.description = source["description"];
	    }
	}
	export class SearchIntent {
	    user_input: string;
	    input_language: string;
	    languages: string[];
	    topics: string[];
	    target_role: string;
	    goals: string[];
	    difficulty: string;
	    direction: string;
	    pushed_after: string;
	    project_size: number;
	    min_stars: number;
	    max_stars: number;

	    static createFrom(source: any = {}) {
	        return new SearchIntent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_input = source["user_input"];
	        this.input_language = source["input_language"];
	        this.languages = source["languages"];
	        this.topics = source["topics"];
	        this.target_role = source["target_role"];
	        this.goals = source["goals"];
	        this.difficulty = source["difficulty"];
	        this.direction = source["direction"];
	        this.pushed_after = source["pushed_after"];
	        this.project_size = source["project_size"];
	        this.min_stars = source["min_stars"];
	        this.max_stars = source["max_stars"];
	    }
	}
	export class DiscoveryResult {
	    intent: SearchIntent;
	    queries: PlannedQuery[];
	    repositories: ScoredRepository[];
	    markdown_report: string;
	    used_live_github: boolean;
	    warnings: string[];

	    static createFrom(source: any = {}) {
	        return new DiscoveryResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.intent = this.convertValues(source["intent"], SearchIntent);
	        this.queries = this.convertValues(source["queries"], PlannedQuery);
	        this.repositories = this.convertValues(source["repositories"], ScoredRepository);
	        this.markdown_report = source["markdown_report"];
	        this.used_live_github = source["used_live_github"];
	        this.warnings = source["warnings"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LLMInsight {
	    readme_summary: string;
	    issue_explanation: string;
	    pr_risk_summary: string;
	    contribution_plan: string;
	    recommendation: string;
	    provider: string;
	    model: string;
	    ai_generated: boolean;
	    generation_warning: string;

	    static createFrom(source: any = {}) {
	        return new LLMInsight(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.readme_summary = source["readme_summary"];
	        this.issue_explanation = source["issue_explanation"];
	        this.pr_risk_summary = source["pr_risk_summary"];
	        this.contribution_plan = source["contribution_plan"];
	        this.recommendation = source["recommendation"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.ai_generated = source["ai_generated"];
	        this.generation_warning = source["generation_warning"];
	    }
	}


	export class RepositoryProfile {
	    repository: Repository;
	    readme_summary: string;
	    structure_summary: string;
	    dependency_summary: string;
	    has_readme: boolean;
	    has_docs: boolean;
	    has_examples: boolean;
	    has_tests: boolean;
	    has_contributing: boolean;
	    good_first_issue_count: number;
	    help_wanted_count: number;

	    static createFrom(source: any = {}) {
	        return new RepositoryProfile(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repository = this.convertValues(source["repository"], Repository);
	        this.readme_summary = source["readme_summary"];
	        this.structure_summary = source["structure_summary"];
	        this.dependency_summary = source["dependency_summary"];
	        this.has_readme = source["has_readme"];
	        this.has_docs = source["has_docs"];
	        this.has_examples = source["has_examples"];
	        this.has_tests = source["has_tests"];
	        this.has_contributing = source["has_contributing"];
	        this.good_first_issue_count = source["good_first_issue_count"];
	        this.help_wanted_count = source["help_wanted_count"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RepositoryAnalysis {
	    repository: Repository;
	    profile: RepositoryProfile;
	    positioning: string;
	    architecture: string;
	    learning_modules: string[];
	    contribution_types: string[];
	    issue_summary: string;
	    pr_summary: string;
	    docs_summary: string;
	    examples_summary: string;
	    tests_summary: string;
	    dependency_files: string[];
	    directory_summary: string;
	    contribution_plan: string;
	    resume_value: string;
	    agent_trace: AgentTraceStep[];
	    llm_insight: LLMInsight;

	    static createFrom(source: any = {}) {
	        return new RepositoryAnalysis(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repository = this.convertValues(source["repository"], Repository);
	        this.profile = this.convertValues(source["profile"], RepositoryProfile);
	        this.positioning = source["positioning"];
	        this.architecture = source["architecture"];
	        this.learning_modules = source["learning_modules"];
	        this.contribution_types = source["contribution_types"];
	        this.issue_summary = source["issue_summary"];
	        this.pr_summary = source["pr_summary"];
	        this.docs_summary = source["docs_summary"];
	        this.examples_summary = source["examples_summary"];
	        this.tests_summary = source["tests_summary"];
	        this.dependency_files = source["dependency_files"];
	        this.directory_summary = source["directory_summary"];
	        this.contribution_plan = source["contribution_plan"];
	        this.resume_value = source["resume_value"];
	        this.agent_trace = this.convertValues(source["agent_trace"], AgentTraceStep);
	        this.llm_insight = this.convertValues(source["llm_insight"], LLMInsight);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class RepositoryQuestionRequest {
	    full_name: string;
	    question: string;

	    static createFrom(source: any = {}) {
	        return new RepositoryQuestionRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.full_name = source["full_name"];
	        this.question = source["question"];
	    }
	}
	export class RepositoryQuestionResponse {
	    full_name: string;
	    question: string;
	    answer: string;
	    provider: string;
	    model: string;
	    ai_generated: boolean;

	    static createFrom(source: any = {}) {
	        return new RepositoryQuestionResponse(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.full_name = source["full_name"];
	        this.question = source["question"];
	        this.answer = source["answer"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.ai_generated = source["ai_generated"];
	    }
	}
	export class ResearchSession {
	    id: number;
	    repository: string;
	    title: string;
	    analysis: RepositoryAnalysis;
	    // Go type: time
	    created_at: any;
	    provider: string;
	    model: string;
	    ai_generated: boolean;
	    trace_step_count: number;

	    static createFrom(source: any = {}) {
	        return new ResearchSession(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repository = source["repository"];
	        this.title = source["title"];
	        this.analysis = this.convertValues(source["analysis"], RepositoryAnalysis);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.ai_generated = source["ai_generated"];
	        this.trace_step_count = source["trace_step_count"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}



	export class SearchRequest {
	    user_input: string;
	    limit: number;
	    input_language: string;
	    languages: string[];
	    topics: string[];
	    target_role: string;
	    difficulty: string;
	    direction: string;
	    pushed_after: string;
	    min_stars: number;
	    max_stars: number;

	    static createFrom(source: any = {}) {
	        return new SearchRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_input = source["user_input"];
	        this.limit = source["limit"];
	        this.input_language = source["input_language"];
	        this.languages = source["languages"];
	        this.topics = source["topics"];
	        this.target_role = source["target_role"];
	        this.difficulty = source["difficulty"];
	        this.direction = source["direction"];
	        this.pushed_after = source["pushed_after"];
	        this.min_stars = source["min_stars"];
	        this.max_stars = source["max_stars"];
	    }
	}

}

export namespace llm {

	export class Provider {
	    name: string;
	    model: string;
	    base_url: string;
	    enabled: boolean;
	    configured: boolean;
	    summary_only: boolean;

	    static createFrom(source: any = {}) {
	        return new Provider(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model = source["model"];
	        this.base_url = source["base_url"];
	        this.enabled = source["enabled"];
	        this.configured = source["configured"];
	        this.summary_only = source["summary_only"];
	    }
	}

}

