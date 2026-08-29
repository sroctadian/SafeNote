export namespace app {
	
	export class DataDirStatus {
	    newPath: string;
	    adopted: boolean;
	    moved: boolean;
	    oldDataBackedUp: boolean;
	    restartRequired: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DataDirStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.newPath = source["newPath"];
	        this.adopted = source["adopted"];
	        this.moved = source["moved"];
	        this.oldDataBackedUp = source["oldDataBackedUp"];
	        this.restartRequired = source["restartRequired"];
	    }
	}
	export class ListNotesResult {
	    notes: domain.NoteCard[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new ListNotesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.notes = this.convertValues(source["notes"], domain.NoteCard);
	        this.total = source["total"];
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
	
	export class BackupNote {
	    id: string;
	    title: string;
	    encrypted_content: string;
	    salt: string;
	    nonce: string;
	    tags: string[];
	    favorite: boolean;
	    pinned: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new BackupNote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.encrypted_content = source["encrypted_content"];
	        this.salt = source["salt"];
	        this.nonce = source["nonce"];
	        this.tags = source["tags"];
	        this.favorite = source["favorite"];
	        this.pinned = source["pinned"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class BackupFile {
	    version: string;
	    // Go type: time
	    created_at: any;
	    checksum: string;
	    notes: BackupNote[];
	
	    static createFrom(source: any = {}) {
	        return new BackupFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.checksum = source["checksum"];
	        this.notes = this.convertValues(source["notes"], BackupNote);
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
	
	export class DecryptedNote {
	    id: string;
	    title: string;
	    content: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new DecryptedNote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.tags = source["tags"];
	    }
	}
	export class NoteCard {
	    id: string;
	    title: string;
	    tags: string[];
	    favorite: boolean;
	    pinned: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new NoteCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.tags = source["tags"];
	        this.favorite = source["favorite"];
	        this.pinned = source["pinned"];
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
	export class Settings {
	    id: string;
	    theme: string;
	    clipboardClearSeconds: number;
	    failedAttemptLimit: number;
	    cooldownSeconds: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.theme = source["theme"];
	        this.clipboardClearSeconds = source["clipboardClearSeconds"];
	        this.failedAttemptLimit = source["failedAttemptLimit"];
	        this.cooldownSeconds = source["cooldownSeconds"];
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

}

export namespace service {
	
	export class RestorePreview {
	    valid: boolean;
	    noteCount: number;
	    duplicateIds: string[];
	    formatVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new RestorePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.noteCount = source["noteCount"];
	        this.duplicateIds = source["duplicateIds"];
	        this.formatVersion = source["formatVersion"];
	    }
	}

}

