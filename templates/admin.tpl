<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <link rel="stylesheet" href="/assets/css/bootstrap.min.css" integrity="sha384-TX8t27EcRE3e/ihU7zmQxVncDAy5uIKz4rEkgIXeMed4M0jlfIDPvg6uqKI2xXr2">
    </head>
    <body>
        <div class="container">
            <div class="row">
                <div class="col-12">

                    <h1 class="mt-4">Admin Panel</h1>
                    <ul class="nav nav-tabs mt-3">
                        <li class="nav-item">
                            <a class="nav-link active" id="settings-tab" data-toggle="tab" href="#settings" role="tab" aria-controls="settings" aria-selected="true">General Settings</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" id="apps-tab" data-toggle="tab" href="#apps" role="tab" aria-controls="apps" aria-selected="false">Apps</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" id="users-tab" data-toggle="tab" href="#users" role="tab" aria-controls="users" aria-selected="false">Users</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" id="groups-tab" data-toggle="tab" href="#groups" role="tab" aria-controls="groups" aria-selected="false">Groups</a>
                        </li>
                    </ul>
                    <div class="tab-content p-3" id="adminTabContent">
                        <div class="tab-pane fade show active" id="settings" role="tabpanel" aria-labelledby="settings-tab">
                            <div class="card">
                                <div class="card-header">R executable</div>
                                <div class="card-body">
                                    <form>
                                        <div class="form-group">
                                            <label for="RscriptPath">Path to Rscript executable</label>
                                            <input type="text" class="form-control" id="RscriptPath">
                                        </div>
                                    </form>
                                    <button class="btn btn-success">Save</button>
                                </div>
                            </div>
                        </div>
                        <div class="tab-pane fade" id="apps" role="tabpanel" aria-labelledby="apps-tab">
                            <div id="accordion">
                                <div class="card">
                                  <div class="card-header" id="headingOne">
                                      <button class="btn btn-link b" data-toggle="collapse" data-target="#collapseOne" aria-expanded="true" aria-controls="collapseOne">
                                        Collapsible Group Item #1
                                      </button>
                                      <button class="btn btn-link float-right" data-toggle="collapse" data-target="#collapseOne" aria-expanded="true" aria-controls="collapseOne">
                                        Collapsible Group Item #1
                                      </button>
                                  </div>
                              
                                  <div id="collapseOne" class="collapse show" aria-labelledby="headingOne" data-parent="#accordion">
                                    <div class="card-body">
                                      Anim pariatur cliche reprehenderit, enim eiusmod high life accusamus terry richardson ad squid. 3 wolf moon officia aute, non cupidatat skateboard dolor brunch. Food truck quinoa nesciunt laborum eiusmod. Brunch 3 wolf moon tempor, sunt aliqua put a bird on it squid single-origin coffee nulla assumenda shoreditch et. Nihil anim keffiyeh helvetica, craft beer labore wes anderson cred nesciunt sapiente ea proident. Ad vegan excepteur butcher vice lomo. Leggings occaecat craft beer farm-to-table, raw denim aesthetic synth nesciunt you probably haven't heard of them accusamus labore sustainable VHS.
                                    </div>
                                  </div>
                                </div>
                                <div class="card">
                                  <div class="card-header" id="headingTwo">
                                    <h5 class="mb-0">
                                      <button class="btn btn-link collapsed" data-toggle="collapse" data-target="#collapseTwo" aria-expanded="false" aria-controls="collapseTwo">
                                        Collapsible Group Item #2
                                      </button>
                                    </h5>
                                  </div>
                                  <div id="collapseTwo" class="collapse" aria-labelledby="headingTwo" data-parent="#accordion">
                                    <div class="card-body">
                                      Anim pariatur cliche reprehenderit, enim eiusmod high life accusamus terry richardson ad squid. 3 wolf moon officia aute, non cupidatat skateboard dolor brunch. Food truck quinoa nesciunt laborum eiusmod. Brunch 3 wolf moon tempor, sunt aliqua put a bird on it squid single-origin coffee nulla assumenda shoreditch et. Nihil anim keffiyeh helvetica, craft beer labore wes anderson cred nesciunt sapiente ea proident. Ad vegan excepteur butcher vice lomo. Leggings occaecat craft beer farm-to-table, raw denim aesthetic synth nesciunt you probably haven't heard of them accusamus labore sustainable VHS.
                                    </div>
                                  </div>
                                </div>
                                <div class="card">
                                  <div class="card-header" id="headingThree">
                                    <h5 class="mb-0">
                                      <button class="btn btn-link collapsed" data-toggle="collapse" data-target="#collapseThree" aria-expanded="false" aria-controls="collapseThree">
                                        Collapsible Group Item #3
                                      </button>
                                    </h5>
                                  </div>
                                  <div id="collapseThree" class="collapse" aria-labelledby="headingThree" data-parent="#accordion">
                                    <div class="card-body">
                                      Anim pariatur cliche reprehenderit, enim eiusmod high life accusamus terry richardson ad squid. 3 wolf moon officia aute, non cupidatat skateboard dolor brunch. Food truck quinoa nesciunt laborum eiusmod. Brunch 3 wolf moon tempor, sunt aliqua put a bird on it squid single-origin coffee nulla assumenda shoreditch et. Nihil anim keffiyeh helvetica, craft beer labore wes anderson cred nesciunt sapiente ea proident. Ad vegan excepteur butcher vice lomo. Leggings occaecat craft beer farm-to-table, raw denim aesthetic synth nesciunt you probably haven't heard of them accusamus labore sustainable VHS.
                                    </div>
                                  </div>
                                </div>
                              </div>
                        </div>
                        <div class="tab-pane fade" id="users" role="tabpanel" aria-labelledby="users-tab">...</div>
                        <div class="tab-pane fade" id="groups" role="tabpanel" aria-labelledby="groups-tab">...</div>
                    </div>
                </div>
            </div>
        </div>
        <script src="/assets/js/jquery-3.5.1.slim.min.js" integrity="sha384-DfXdz2htPH0lsSSs5nCTpuj/zy4C+OGpamoFVy38MVBnE+IbbVYUew+OrCXaRkfj"></script>
        <script src="/assets/js/bootstrap.bundle.min.js" integrity="sha384-ho+j7jyWK8fNQe+A12Hb8AhRq26LrZ/JpcUGGOn+Y7RsweNrtN/tE3MoK7ZeZDyx"></script>
    </body>
</html>

