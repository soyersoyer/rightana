import { Injectable } from '@angular/core';
import { HttpClient, HttpEvent, HttpErrorResponse, HttpInterceptor, HttpHandler, HttpRequest } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { Router } from '@angular/router';
import 'rxjs/add/operator/do';
import 'rxjs/add/observable/of';

import { ToastyService } from './toasty/toasty.module';

export class ServerConfig {
  enable_registration: boolean;
  tracking_id: string;
  server_announce: string;
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
  email_verified: boolean;
}

export class UserCreate {
  name: string;
  email: string;
  password: string;
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
  user: string;
  name: string;
  session_count: number;
  session_percent: number;
  session_sums: BucketSum[];
  pageview_count: number;
  pageview_percent: number;
  pageview_sums: BucketSum[];
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

export class Backup {
  id: string;
  dir: string;
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
  user: string;
  is_admin: boolean;

  constructor(
    private router: Router,
  ) {
    this.token = localStorage.getItem('token');
    this.user = localStorage.getItem('user');
    this.is_admin = localStorage.getItem('is_admin') === 'true';
  }

  set(token: string, user: string, is_admin: boolean) {
    this.token = token;
    this.user = user;
    this.is_admin = is_admin;
    localStorage.setItem('token', token);
    localStorage.setItem('user', user);
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

  config: ServerConfig;
  getConfig(): Observable<ServerConfig> {
    if(!this.config) {
      return this.http.get<ServerConfig>('/api/config').do(config => this.config = config);
    }
    return Observable.of(this.config);
  }

  createUser(formData: any): Observable<User> {
    return this.http.post<User>('/api/users', formData);
  }

  createAuthToken(name_or_email: string, password: string): Observable<AuthToken> {
    return this.http.post<AuthToken>('/api/authtokens', {name_or_email, password});
  }

  deleteAuthToken(authToken: string): Observable<any> {
    return this.http.delete(`api/authtokens/${authToken}`);
  }

  createCollection(user: string, formData: any): Observable<Collection> {
   return this.http.post<Collection>(`/api/users/${user}/collections/create-new`, formData);
  }

  getCollectionSummaries(user: string, timezone: string): Observable<CollectionSummary[]> {
    return this.http.post<CollectionSummary[]>(`/api/users/${user}/collections`, {timezone});
  }

  getCollection(user: string, collectionName: string): Observable<Collection> {
   return this.http.get<Collection>(`/api/users/${user}/collections/${collectionName}`);
  }

  saveCollection(user: string, collectionName: string, formData: any): Observable<Collection> {
    return this.http.put<Collection>(`/api/users/${user}/collections/${collectionName}`, formData);
  }

  deleteCollection(user: string, collectionName: string): Observable<any> {
    return this.http.delete(`/api/users/${user}/collections/${collectionName}`);
  }

  getCollectionShards(user: string, collectionName: string): Observable<Shard[]> {
    return this.http.get<Shard[]>(`/api/users/${user}/collections/${collectionName}/shards`);
  }

  deleteCollectionShard(user: string, collectionName: string, shardId: string): Observable<any> {
    return this.http.delete(`/api/users/${user}/collections/${collectionName}/shards/${shardId}`);
  }

  getTeammates(user: string, collectionName: string): Observable<Teammate[]> {
    return this.http.get<Teammate[]>(`/api/users/${user}/collections/${collectionName}/teammates`);
  }

  addTeammate(user: string, collectionName: string, email: string): Observable<Teammate> {
    return this.http.post<Teammate>(`/api/users/${user}/collections/${collectionName}/teammates`, {email});
  }

  removeTeammate(user: string, collectionName: string, email: string): Observable<Teammate> {
    return this.http.delete<Teammate>(`/api/users/${user}/collections/${collectionName}/teammates/${email}`);
  }

  getCollectionData(user: string, collectionName: string, from: Date, to: Date, bucket: string, timezone: string, filter: any): Observable<CollectionData> {
    return this.http.post<CollectionData>(`/api/users/${user}/collections/${collectionName}/data`, {from, to, bucket, timezone, filter});
  }

  getCollectionStatData(user: string, collectionName: string, from: Date, to: Date, filter: any): Observable<CollectionSumData> {
    return this.http.post<CollectionSumData>(`/api/users/${user}/collections/${collectionName}/stat`, {from, to, filter});
  }

  getSessions(user: string, collectionName: string, from: Date, to: Date, filter: any): Observable<Session[]> {
   return this.http.post<Session[]>(`/api/users/${user}/collections/${collectionName}/sessions`, {from, to, filter});
  }

  getPageviews(user: string, collectionName: string, sessionKey: string): Observable<Pageview[]> {
   return this.http.post<Pageview[]>(`/api/users/${user}/collections/${collectionName}/pageviews`, {session_key: sessionKey});
  }

  getUserInfo(name: string): Observable<UserInfo> {
    return this.http.get<UserInfo>(`/api/users/${name}`);
  }

  updateUserPassword(user: string, currentPassword: string, password: string): Observable<any> {
    return this.http.patch(`/api/users/${user}/settings/password`, {currentPassword, password});
  }

  deleteUser(user: string, password: string): Observable<any> {
   return this.http.post(`/api/users/${user}/settings/delete`, {password});
  }

  sendVerifyEmail(user: string): Observable<any> {
    return this.http.post(`/api/users/${user}/settings/send-verify-email`, {});
  }

  verifyEmail(user: string, verificationKey: string): Observable<any> {
    return this.http.post(`/api/users/${user}/verify-email`, {verificationKey});
  }

  sendResetPassword(email: string): Observable<any> {
    return this.http.post(`/api/users/send-reset-password`, {email});
  }

  resetPassword(user: string, resetKey: string, password: string): Observable<any> {
    return this.http.post(`/api/users/${user}/reset-password`, {resetKey, password});
  }

  getUsers(): Observable<UserInfo[]> {
    return this.http.get<UserInfo[]>(`/api/admin/users`);
  }

  createUserAdmin(user: UserCreate): Observable<string> {
    return this.http.post<string>(`/api/admin/users`, user);
  }

  updateUser(name: string, user: UserUpdate): Observable<string> {
    return this.http.patch<string>(`/api/admin/users/${name}`, user);
  }

  deleteUserAdmin(user: string): Observable<any> {
    return this.http.delete(`/api/admin/users/${user}`);
   }

  getCollections(): Observable<CollectionInfo[]> {
    return this.http.get<CollectionInfo[]>(`/api/admin/collections`);
  }

  getBackups(): Observable<Backup[]> {
    return this.http.get<Backup[]>(`/api/backups`);
  }

  runBackup(id: string): Observable<any> {
    return this.http.get<any>(`/api/backups/${id}/run`);
  }

}
