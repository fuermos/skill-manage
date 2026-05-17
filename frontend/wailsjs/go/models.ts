export namespace main {
	
	export class AppConfig {
	    server: string;
	    token: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.token = source["token"];
	    }
	}
	export class AppStatus {
	    revision: number;
	    updated: string;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.revision = source["revision"];
	        this.updated = source["updated"];
	    }
	}
	export class DiffItem {
	    path: string;
	    tool: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.tool = source["tool"];
	        this.status = source["status"];
	    }
	}
	export class PullChange {
	    path: string;
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new PullChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.action = source["action"];
	    }
	}
	export class PullResult {
	    changes: PullChange[];
	
	    static createFrom(source: any = {}) {
	        return new PullResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.changes = this.convertValues(source["changes"], PullChange);
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
	export class PushResult {
	    applied: number;
	    new_revision: number;
	
	    static createFrom(source: any = {}) {
	        return new PushResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applied = source["applied"];
	        this.new_revision = source["new_revision"];
	    }
	}
	export class RecItem {
	    to_skill_id: string;
	    score: number;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new RecItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.to_skill_id = source["to_skill_id"];
	        this.score = source["score"];
	        this.reason = source["reason"];
	    }
	}
	export class SkillInfo {
	    name: string;
	    display_name: string;
	    tool: string;
	    category: string;
	    size: number;
	    summary: string;
	    description: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.tool = source["tool"];
	        this.category = source["category"];
	        this.size = source["size"];
	        this.summary = source["summary"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	    }
	}
	export class ToolInfo {
	    name: string;
	    display: string;
	    enabled: boolean;
	    installed: boolean;
	    files: number;
	
	    static createFrom(source: any = {}) {
	        return new ToolInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display = source["display"];
	        this.enabled = source["enabled"];
	        this.installed = source["installed"];
	        this.files = source["files"];
	    }
	}

}

