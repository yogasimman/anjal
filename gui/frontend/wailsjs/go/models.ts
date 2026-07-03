export namespace models {
	
	export class Auth {
	    Type: string;
	    Params: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Auth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Params = source["Params"];
	    }
	}
	export class APIRequest {
	    ID: string;
	    Title: string;
	    Method: string;
	    URL: string;
	    QueryParams: Record<string, string>;
	    Headers: Record<string, string>;
	    Auth?: Auth;
	    Body: string;
	
	    static createFrom(source: any = {}) {
	        return new APIRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Title = source["Title"];
	        this.Method = source["Method"];
	        this.URL = source["URL"];
	        this.QueryParams = source["QueryParams"];
	        this.Headers = source["Headers"];
	        this.Auth = this.convertValues(source["Auth"], Auth);
	        this.Body = source["Body"];
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
	export class APIResponse {
	    StatusCode: number;
	    Status: string;
	    Body: string;
	    Latency: number;
	    Headers: Record<string, Array<string>>;
	    ContentType: string;
	
	    static createFrom(source: any = {}) {
	        return new APIResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.StatusCode = source["StatusCode"];
	        this.Status = source["Status"];
	        this.Body = source["Body"];
	        this.Latency = source["Latency"];
	        this.Headers = source["Headers"];
	        this.ContentType = source["ContentType"];
	    }
	}
	
	export class Collection {
	    Name: string;
	    FilePath: string;
	    Auth?: Auth;
	    Requests: APIRequest[];
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.FilePath = source["FilePath"];
	        this.Auth = this.convertValues(source["Auth"], Auth);
	        this.Requests = this.convertValues(source["Requests"], APIRequest);
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

