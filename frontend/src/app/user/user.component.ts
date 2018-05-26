import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

@Component({
  selector: 'rana-user',
  templateUrl: './user.component.html',
})
export class UserComponent implements OnInit {
  user: string;

  constructor(
    private route: ActivatedRoute,
  ) { }

  ngOnInit() {
    this.route.params.forEach((params: Params) => {
      this.user = params['user'];
    });
  }

}
