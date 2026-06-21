export namespace domain {
	
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
	    languages: string[];
	    topics: string[];
	    target_role: string;
	    goals: string[];
	    difficulty: string;
	    project_size: number;
	    min_stars: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchIntent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_input = source["user_input"];
	        this.languages = source["languages"];
	        this.topics = source["topics"];
	        this.target_role = source["target_role"];
	        this.goals = source["goals"];
	        this.difficulty = source["difficulty"];
	        this.project_size = source["project_size"];
	        this.min_stars = source["min_stars"];
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
	
	
	
	

}

