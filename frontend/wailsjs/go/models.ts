export namespace ssh {
	
	export class Client {
	    Conn: any;
	
	    static createFrom(source: any = {}) {
	        return new Client(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Conn = source["Conn"];
	    }
	}

}

