export namespace models {
	
	export class AIModel {
	    id: string;
	    name: string;
	    description: string;
	    maxTokens: number;
	    supportsVision: boolean;
	    provider: string;
	
	    static createFrom(source: any = {}) {
	        return new AIModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.maxTokens = source["maxTokens"];
	        this.supportsVision = source["supportsVision"];
	        this.provider = source["provider"];
	    }
	}
	export class AIResult {
	    contentType: string;
	    title: string;
	    keywords: string[];
	    quality: number;
	    description: string;
	    category: string;
	    processed: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contentType = source["contentType"];
	        this.title = source["title"];
	        this.keywords = source["keywords"];
	        this.quality = source["quality"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.processed = source["processed"];
	        this.error = source["error"];
	    }
	}
	export class AppSettings {
	    id: string;
	    tempDirectory: string;
	    aiProvider: string;
	    aiModel: string;
	    aiApiKey: string;
	    aiBaseUrl: string;
	    maxConcurrentJobs: number;
	    aiTimeout: number;
	    aiMaxTokens: number;
	    thumbnailSize: number;
	    language: string;
	    aiPrompts: Record<string, string>;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tempDirectory = source["tempDirectory"];
	        this.aiProvider = source["aiProvider"];
	        this.aiModel = source["aiModel"];
	        this.aiApiKey = source["aiApiKey"];
	        this.aiBaseUrl = source["aiBaseUrl"];
	        this.maxConcurrentJobs = source["maxConcurrentJobs"];
	        this.aiTimeout = source["aiTimeout"];
	        this.aiMaxTokens = source["aiMaxTokens"];
	        this.thumbnailSize = source["thumbnailSize"];
	        this.language = source["language"];
	        this.aiPrompts = source["aiPrompts"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class BatchStats {
	    total: number;
	    processed: number;
	    approved: number;
	    rejected: number;
	
	    static createFrom(source: any = {}) {
	        return new BatchStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.processed = source["processed"];
	        this.approved = source["approved"];
	        this.rejected = source["rejected"];
	    }
	}
	export class PhotoProcessInfo {
	    id: string;
	    fileName: string;
	    status: string;
	    progress: number;
	    step: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new PhotoProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fileName = source["fileName"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.step = source["step"];
	        this.error = source["error"];
	    }
	}
	export class BatchStatus {
	    batchId: string;
	    type: string;
	    description: string;
	    totalPhotos: number;
	    processedPhotos: number;
	    status: string;
	    progress: number;
	    currentPhoto?: string;
	    currentStep?: string;
	    photos?: PhotoProcessInfo[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.batchId = source["batchId"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.totalPhotos = source["totalPhotos"];
	        this.processedPhotos = source["processedPhotos"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.currentPhoto = source["currentPhoto"];
	        this.currentStep = source["currentStep"];
	        this.photos = this.convertValues(source["photos"], PhotoProcessInfo);
	        this.error = source["error"];
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
	export class ConnectionConfig {
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    path: string;
	    apiKey?: string;
	    apiUrl?: string;
	    timeout?: number;
	    useSSL?: boolean;
	    passive?: boolean;
	    encryption?: string;
	    verifyCert?: boolean;
	    headers?: Record<string, string>;
	    params?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.path = source["path"];
	        this.apiKey = source["apiKey"];
	        this.apiUrl = source["apiUrl"];
	        this.timeout = source["timeout"];
	        this.useSSL = source["useSSL"];
	        this.passive = source["passive"];
	        this.encryption = source["encryption"];
	        this.verifyCert = source["verifyCert"];
	        this.headers = source["headers"];
	        this.params = source["params"];
	    }
	}
	export class EventLog {
	    id: string;
	    batchId: string;
	    photoId?: string;
	    eventType: string;
	    status: string;
	    message: string;
	    details?: string;
	    progress?: number;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EventLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.batchId = source["batchId"];
	        this.photoId = source["photoId"];
	        this.eventType = source["eventType"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.details = source["details"];
	        this.progress = source["progress"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class Photo {
	    id: string;
	    batchId: string;
	    contentType: string;
	    originalPath: string;
	    thumbnailPath: string;
	    fileName: string;
	    fileSize: number;
	    exifData: Record<string, string>;
	    aiResult?: AIResult;
	    uploadStatus: Record<string, string>;
	    status: string;
	    selectedForUpload: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new Photo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.batchId = source["batchId"];
	        this.contentType = source["contentType"];
	        this.originalPath = source["originalPath"];
	        this.thumbnailPath = source["thumbnailPath"];
	        this.fileName = source["fileName"];
	        this.fileSize = source["fileSize"];
	        this.exifData = source["exifData"];
	        this.aiResult = this.convertValues(source["aiResult"], AIResult);
	        this.uploadStatus = source["uploadStatus"];
	        this.status = source["status"];
	        this.selectedForUpload = source["selectedForUpload"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class PhotoBatch {
	    id: string;
	    type: string;
	    description: string;
	    folderPath: string;
	    photos: Photo[];
	    photosStats?: BatchStats;
	    status: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PhotoBatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.folderPath = source["folderPath"];
	        this.photos = this.convertValues(source["photos"], Photo);
	        this.photosStats = this.convertValues(source["photosStats"], BatchStats);
	        this.status = source["status"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class PhotoFile {
	    name: string;
	    path: string;
	    size: number;
	    extension: string;
	    isValid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PhotoFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.extension = source["extension"];
	        this.isValid = source["isValid"];
	    }
	}
	
	export class ProcessingProgress {
	    batchId: string;
	    totalPhotos: number;
	    currentStep: string;
	    currentPhoto: string;
	    photoProgress: number;
	    overallProgress: number;
	    status: string;
	    recentEvents: EventLog[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessingProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.batchId = source["batchId"];
	        this.totalPhotos = source["totalPhotos"];
	        this.currentStep = source["currentStep"];
	        this.currentPhoto = source["currentPhoto"];
	        this.photoProgress = source["photoProgress"];
	        this.overallProgress = source["overallProgress"];
	        this.status = source["status"];
	        this.recentEvents = this.convertValues(source["recentEvents"], EventLog);
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
	export class StockConfig {
	    id: string;
	    name: string;
	    type: string;
	    supportedTypes: string[];
	    uploadMethod: string;
	    connection: ConnectionConfig;
	    prompts: Record<string, string>;
	    settings: Record<string, any>;
	    active: boolean;
	    modulePath: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new StockConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.supportedTypes = source["supportedTypes"];
	        this.uploadMethod = source["uploadMethod"];
	        this.connection = this.convertValues(source["connection"], ConnectionConfig);
	        this.prompts = source["prompts"];
	        this.settings = source["settings"];
	        this.active = source["active"];
	        this.modulePath = source["modulePath"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class UploaderInfo {
	    name: string;
	    version: string;
	    description: string;
	    author: string;
	    type: string;
	    website?: string;
	
	    static createFrom(source: any = {}) {
	        return new UploaderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.author = source["author"];
	        this.type = source["type"];
	        this.website = source["website"];
	    }
	}

}

