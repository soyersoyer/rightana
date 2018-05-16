import { Injectable } from '@angular/core';
import { HttpClient, HttpEvent, HttpErrorResponse, HttpInterceptor, HttpHandler, HttpRequest } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { Router } from '@angular/router';
import 'rxjs/add/operator/do';

import { ToastyService } from 'ng2-toasty';

export class ServerConfig {
  enable_registration: boolean;
}

export class User {
  email: string;
  password: string;
}

export class UserInfo {
  id: number
  email: string;
  name: string;
  created: number;
  is_admin: boolean;
  collection_count: number;
}

export class UserUpdate {
  password: string;
  is_admin: string;
}

export class AuthToken {
  id: string;
  user_info: UserInfo;
}

export class Collection {
  id: string;
  name: string;
}

export class CollectionInfo {
  id: string;
  name: string;
  owner_name: string;
  created: number;
  teammate_count: number;
}

export class CollectionSummary {
  id: string;
  name: string;
  pageview_count: number;
  pageview_percent: number;
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
  owner_name: string;
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
  query_string: string;
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
  name: string;
  is_admin: boolean;

  constructor(
    private router: Router,
  ) {
    this.token = localStorage.getItem('token');
    this.name = localStorage.getItem('name');
    this.is_admin = localStorage.getItem('is_admin') === 'true';
  }

  set(token: string, name: string, is_admin: boolean) {
    this.token = token;
    this.name = name;
    this.is_admin = is_admin;
    localStorage.setItem('token', token);
    localStorage.setItem('name', name);
    localStorage.setItem('is_admin', is_admin?'true':'false');
  }

  unset() {
    this.set('', '', false);
  }

  get loggedIn(): boolean {
    return this.token && this.token !== '';
  }

  get isAdmin(): boolean {
    return this.loggedIn && this.is_admin;
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

  getConfig(): Observable<ServerConfig> {
    return this.http.get<ServerConfig>('/api/config');
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

  getCollectionSummaries(): Observable<CollectionSummary[]> {
    return this.http.get<CollectionSummary[]>('/api/collections');
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

  updateUserPassword(name: string, currentPassword: string, password: string): Observable<any> {
    return this.http.patch(`/api/users/${name}/settings/password`, {currentPassword, password});
  }

  deleteUser(name: string, password: string): Observable<any> {
   return this.http.post(`/api/users/${name}/settings/delete`, {password});
  }

  getUsers(): Observable<UserInfo[]> {
    return this.http.get<UserInfo[]>(`/api/admin/users`);
  }

  getUserInfo(email: string): Observable<UserInfo> {
    return this.http.get<UserInfo>(`/api/admin/users/${email}`);
  }

  updateUser(email: string, user: UserUpdate): Observable<string> {
    return this.http.patch<string>(`/api/admin/users/${email}`, user);
  }

  getCollections(): Observable<CollectionInfo[]> {
    return this.http.get<CollectionInfo[]>(`/api/admin/collections`);
  }

}
