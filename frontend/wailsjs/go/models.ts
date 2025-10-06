export namespace pe {
	
	export class ExportFunction {
	    ordinal: number;
	    function_rva: number;
	    name_ordinal: number;
	    name_rva: number;
	    name: string;
	    forwarder: string;
	    forwarder_rva: number;
	
	    static createFrom(source: any = {}) {
	        return new ExportFunction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ordinal = source["ordinal"];
	        this.function_rva = source["function_rva"];
	        this.name_ordinal = source["name_ordinal"];
	        this.name_rva = source["name_rva"];
	        this.name = source["name"];
	        this.forwarder = source["forwarder"];
	        this.forwarder_rva = source["forwarder_rva"];
	    }
	}

}

