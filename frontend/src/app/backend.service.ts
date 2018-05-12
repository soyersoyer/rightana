import { Injectable } from '@angular/core';
import { HttpClient, HttpEvent, HttpErrorResponse, HttpInterceptor, HttpHandler, HttpRequest } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { Router } from '@angular/router';
import 'rxjs/add/operator/do';

import { ToastyService } from 'ng2-toasty';

export class K20Config {
  enable_registration: boolean;
}

export class User {
  email: string;
  password: string;
}

export class AuthToken {
  id: string;
}

export class Collection {
  id: string;
  owner_email: number;
  name: string;
}

export class Total {
  count: number;
  diff_percent: number;
}

export class BucketSum {
  bucket: number;
  count: number;
}

export class CollectionData {
  id: string;
  name: string;
  owner_email: string;
  session_sums: BucketSum[];
  pageview_sums: BucketSum[];
}

export class Teammate {
  email: string;
}

export class CollectionSumData {
  session_total: Total;
  pageview_total: Total;
  avg_session_length: Total;
  page_sums: any;
  referrer_sums: any;
  hostname_sums: any;
  browser_name_sums: any;
  browser_language_sums: any;
  screen_resolution_sums: any;
  window_resolution_sums: any;
  country_code_sums: any;
  city_sums: any;
}

export class Session {
  key: string;
  device_os: string;
  browser_name: string;
  browser_version: string;
  browser_language: string;
  screen_resolution: string;
  window_resolution: string;
  device_type: string;
  country_code: string;
  city: string;
  user_agent: string;
  begin: Date;
  duration: number;
}

export class Pageview {
  time: Date;
  path: string;
  referrer_url: string;
}

export class Shard {
  id: string;
  size: number;
}

@Injectable()
export class ErrorInterceptor implements HttpInterceptor {
  constructor(
    private toasty: ToastyService,
    private auth: AuthService,
  ) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next
      .handle(req)
      .do(null, event => {
        if (event instanceof HttpErrorResponse) {
          this.toasty.error(event.error);
          if (event.error.indexOf("Authtoken expired") == 0) {
            /* we should logout */
            this.auth.unset();
            this.auth.goToLogin();
          }
        }
      });
  }
}

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private auth: AuthService) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    if (this.auth.token) {
      const authReq = req.clone({setHeaders: {Authorization: this.auth.token}});
      return next.handle(authReq);
    }
    return next.handle(req);
  }
}

@Injectable()
export class AuthService {
  token: string;
  email: string;

  constructor(
    private router: Router,
  ) {
    this.token = localStorage.getItem('token');
    this.email = localStorage.getItem('email');
  }

  set(token: string, email: string) {
    this.token = token;
    this.email = email;
    localStorage.setItem('token', token);
    localStorage.setItem('email', email);
  }

  unset() {
    this.set('', '');
  }

  get loggedIn(): boolean {
    return this.token && this.token !== '';
  }

  goToLogin() {
    this.router.navigateByUrl("/login");
  }
}

@Injectable()
export class BackendService {

  constructor(
    private http: HttpClient
  ) {}

  getConfig(): Observable<K20Config> {
    return this.http.get<K20Config>('/api/config');
  }

  createUser(formData: any): Observable<User> {
    return this.http.post<User>('/api/users', JSON.stringify(formData));
  }

  createAuthToken(formData: any): Observable<AuthToken> {
    return this.http.post<AuthToken>('/api/authtokens', JSON.stringify(formData));
  }

  deleteAuthToken(authToken: string): Observable<any> {
    return this.http.delete(`api/authtokens/${authToken}`);
  }

  createCollection(formData: any): Observable<Collection> {
   return this.http.post<Collection>('/api/collections', JSON.stringify(formData));
  }

  getCollections(): Observable<Collection[]> {
    return this.http.get<Collection[]>('/api/collections');
  }

  getCollection(collectionId: string): Observable<Collection> {
   return this.http.get<Collection>(`/api/collections/${collectionId}`);
  }

  saveCollection(formData: any): Observable<Collection> {
    return this.http.put<Collection>(`/api/collections/${formData.id}`, JSON.stringify(formData));
  }

  deleteCollection(collectionId: string): Observable<any> {
    return this.http.delete(`/api/collections/${collectionId}`);
  }

  getCollectionShards(collectionId: string): Observable<Shard[]> {
    return this.http.get<Shard[]>(`/api/collections/${collectionId}/shards`);
  }

  deleteCollectionShard(collectionId: string, shardId: string): Observable<any> {
    return this.http.delete(`/api/collections/${collectionId}/shards/${shardId}`);
  }

  getTeammates(collectionId: string): Observable<Teammate[]> {
    return this.http.get<Teammate[]>(`/api/collections/${collectionId}/teammates`);
  }

  addTeammate(collectionId: string, email: string): Observable<Teammate> {
    return this.http.post<Teammate>(`/api/collections/${collectionId}/teammates`, JSON.stringify({email}));
  }

  removeTeammate(collectionId: string, email: string): Observable<Teammate> {
    return this.http.delete<Teammate>(`/api/collections/${collectionId}/teammates/${email}`);
  }

  getCollectionData(collectionId: string, from: Date, to: Date, bucket: string, timezone: string, filter: any): Observable<CollectionData> {
    return this.http.post<CollectionData>(`/api/collections/${collectionId}/data`, JSON.stringify({from, to, bucket, timezone, filter}));
  }

  getCollectionStatData(collectionId: string, from: Date, to: Date, filter: any): Observable<CollectionSumData> {
    return this.http.post<CollectionSumData>(`/api/collections/${collectionId}/stat`, JSON.stringify({from, to, filter}));
  }

  getSessions(collectionId: string, from: Date, to: Date, filter: any): Observable<Session[]> {
   return this.http.post<Session[]>(`/api/collections/${collectionId}/sessions`, JSON.stringify({from, to, filter}));
  }

  getPageviews(collectionId: string, sessionKey: string): Observable<Pageview[]> {
   return this.http.post<Pageview[]>(`/api/collections/${collectionId}/pageviews`, JSON.stringify({session_key: sessionKey}));
  }

  updateUserPassword(email: string, currentPassword: string, password: string): Observable<any> {
    return this.http.patch(`/api/users/${email}/password`, {currentPassword, password});
  }

  deleteUser(email: string, password: string): Observable<any> {
   return this.http.post(`/api/users/${email}/delete`, {password});
  }

}
