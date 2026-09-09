export namespace config {
	
	export class Settings {
	    user: string;
	    outputFolder: string;
	    watermarkPosition: string;
	    watermarkX: number;
	    watermarkY: number;
	    watermarkWidth: number;
	    fontFamily: string;
	    fontSize: number;
	    useManualLocation: boolean;
	    manualLatitude: number;
	    manualLongitude: number;
	    manualAddress: string;
	    reverseGeocode: boolean;
	    cameraId: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user = source["user"];
	        this.outputFolder = source["outputFolder"];
	        this.watermarkPosition = source["watermarkPosition"];
	        this.watermarkX = source["watermarkX"];
	        this.watermarkY = source["watermarkY"];
	        this.watermarkWidth = source["watermarkWidth"];
	        this.fontFamily = source["fontFamily"];
	        this.fontSize = source["fontSize"];
	        this.useManualLocation = source["useManualLocation"];
	        this.manualLatitude = source["manualLatitude"];
	        this.manualLongitude = source["manualLongitude"];
	        this.manualAddress = source["manualAddress"];
	        this.reverseGeocode = source["reverseGeocode"];
	        this.cameraId = source["cameraId"];
	    }
	}

}

export namespace main {
	
	export class CurrentLocation {
	    latitude: number;
	    longitude: number;
	    accuracy?: number;
	
	    static createFrom(source: any = {}) {
	        return new CurrentLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latitude = source["latitude"];
	        this.longitude = source["longitude"];
	        this.accuracy = source["accuracy"];
	    }
	}
	export class LocationDetails {
	    address: string;
	    roadClue: string;
	
	    static createFrom(source: any = {}) {
	        return new LocationDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.roadClue = source["roadClue"];
	    }
	}
	export class SaveRequest {
	    jpegDataUrl: string;
	    capturedAt: string;
	    user: string;
	    latitude: number;
	    longitude: number;
	    accuracy?: number;
	    location: string;
	    locationSource: string;
	    outputFolder: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jpegDataUrl = source["jpegDataUrl"];
	        this.capturedAt = source["capturedAt"];
	        this.user = source["user"];
	        this.latitude = source["latitude"];
	        this.longitude = source["longitude"];
	        this.accuracy = source["accuracy"];
	        this.location = source["location"];
	        this.locationSource = source["locationSource"];
	        this.outputFolder = source["outputFolder"];
	    }
	}

}

export namespace photo {
	
	export class Item {
	    name: string;
	    path: string;
	    // Go type: time
	    modifiedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
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

