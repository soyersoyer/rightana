import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { ReactiveFormsModule } from '@angular/forms';
import { RouterModule, Routes } from '@angular/router';

import { ToastyModule, ToastyService } from 'ng2-toasty';

import { AppComponent } from './app.component';
import { AuthService, BackendService, AuthInterceptor, ErrorInterceptor } from './backend.service';
import { LoginComponent } from './login/login.component';
import { HomeComponent } from './home/home.component';
import { RegistrationComponent } from './registration/registration.component';
import { CollectionDashboardComponent } from './collection-dashboard/dashboard.component';
import { CollectionComponent } from './collection/collection.component';
import { CollectionCreateComponent } from './collection-create/create.component';
import { LogoutComponent } from './logout/logout.component';
import { CollectionSettingsComponent } from './collection-settings/settings.component';
import { TeammatesComponent } from './collection-settings/teammates.component';
import { CollectionTrackingComponent } from './collection-tracking/tracking.component';
import { SessionComponent } from './session/session.component';
import { CollectionStatComponent } from './collection-stat/stat.component';
import { TableSumComponent } from './collection-stat/table-sum.component';

import { SettingsComponent } from './settings/settings.component';
import { ChangePasswordComponent } from './settings/change-password/change-password.component';
import { DeleteAccountComponent } from './settings/delete-account/delete-account.component';

import { ChartComponent } from './chart/chart.component';

import { ColorPercentComponent } from './utils/color-percent.component';
import { BytesPipe } from './utils/bytes.pipe';

import { AdminComponent } from './admin/admin.component';
import { AdminUsersComponent } from './admin-users/admin-users.component';
import { AdminCollectionsComponent } from './admin-collections/admin-collections.component';
import { AdminUsersEditComponent } from './admin-users-edit/admin-users-edit.component';
import { UserComponent } from './user/user.component';



const routes: Routes = [
  { path: '', component: HomeComponent},
  { path: 'login', component: LoginComponent},
  { path: 'logout', component: LogoutComponent},
  { path: 'registration', component: RegistrationComponent},
  { path: 'settings', component: SettingsComponent, children: [
    { path: '', redirectTo: 'change-password', pathMatch: 'full'},
    { path: 'change-password', component: ChangePasswordComponent},
    { path: 'delete-account', component: DeleteAccountComponent},
  ]},
  { path: 'admin', component: AdminComponent, children: [
    { path: '', redirectTo: 'users', pathMatch: 'full'},
    { path: 'users', component: AdminUsersComponent},
    { path: 'users/:email', component: AdminUsersEditComponent},
    { path: 'collections', component: AdminCollectionsComponent},
  ]},
  { path: ':user', component: UserComponent, children: [
    { path: "", component: CollectionComponent},
    { path: 'create', component: CollectionCreateComponent},
    { path: ':collectionId', component: CollectionDashboardComponent, children: [
      { path: '', redirectTo: 'statistics', pathMatch: 'full'},
      { path: 'statistics', component: CollectionStatComponent},
      { path: 'sessions', component: SessionComponent},
      { path: 'settings', component: CollectionSettingsComponent},
    ]},
  ]},

];

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    HomeComponent,
    RegistrationComponent,
    CollectionComponent,
    CollectionCreateComponent,
    CollectionDashboardComponent,
    LogoutComponent,
    CollectionSettingsComponent,
    TeammatesComponent,
    CollectionTrackingComponent,
    SessionComponent,
    CollectionStatComponent,
    TableSumComponent,
    SettingsComponent,
    ChangePasswordComponent,
    DeleteAccountComponent,
    ChartComponent,
    ColorPercentComponent,
    BytesPipe,
    AdminComponent,
    AdminUsersComponent,
    AdminCollectionsComponent,
    AdminUsersEditComponent,
    UserComponent,
  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    ReactiveFormsModule,
    RouterModule.forRoot(routes),
    ToastyModule.forRoot(),
  ],
  providers: [
    BackendService,
    AuthService,
    { provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true, deps: [AuthService]},
    { provide: HTTP_INTERCEPTORS, useClass: ErrorInterceptor, multi: true, deps: [ToastyService, AuthService] },
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
