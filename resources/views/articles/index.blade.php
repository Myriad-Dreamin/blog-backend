@extends('layouts.mainview')

@section('main_content')
    @foreach($articles as $article)
        {{ $article }}
    @endforeach
@stop
