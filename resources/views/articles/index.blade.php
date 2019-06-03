@extends('layouts.mainview')

@section('extension_files')
    <link rel="stylesheet" href="{{ URL::asset('css/articles/article_index_box.css') }}">
    <link rel="stylesheet" href="{{ URL::asset('css/articles/article_index_body.css') }}">
@stop

@section('extension_metas')
    <title>Article</title>
@stop


@section('main_content')
    <div class="article_index_body"> 
        <div style="height: 50px; width:100%;"></div>
        @foreach($articles as $article)
            <div class="article_index_box">
                <div class="title">
                    <a href="{{url('articles', $article->id)}}">{{ $article->title }}</a>
                </div>
                <div class="tagbox">
                    <a class="category">Category: {{ $article->category }}</a>
                </div>
                <div class="tagbox">
                    <a class="timetag">Publish at:{{ $article->published_at }}</a>
                </div>
                <div class="intro">
                    <a>{{ $article->intro }}</a>
                </div>
            </div>
        @endforeach
    </div>
@stop
