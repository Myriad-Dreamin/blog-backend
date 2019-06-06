@extends('layouts.mainview')

@section('extension_files')
    <link rel="stylesheet" href="{{ URL::asset('css/extension/chocolate.css') }}">
@stop

@section('extension_metas')
    <title>Chocolate</title>
@stop

@section('main_content')
<div class="chocolate_index_body"> 
        <div style="height: 50px; width:100%;"></div>
        @foreach($chocos as $choco)
            <div class="index_box">
                <div class="title">
                    <a href="{{ $choco->ref }}">{{ $choco->title }}</a>
                </div>
                <div class="tagbox">
                    <a class="category">Category: {{ $choco->category }}</a>
                </div>
                <div class="intro">
                    <a>{{ $choco->intro }}</a>
                </div>
            </div>
        @endforeach
    </div>
@stop
