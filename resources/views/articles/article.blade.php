@extends('layouts.mainview')

@section('extension_files')
    <script type="text/x-mathjax-config">
    MathJax.Hub.Config({
        showProcessingMessages: false,
        messageStyle: "none",
        extensions: ["tex2jax.js"],
        jax: ["input/TeX", "output/HTML-CSS"],
        tex2jax: {
            inlineMath:  [ ["$", "$"] ],
            displayMath: [ ["$$","$$"] ],
            skipTags: ['script', 'noscript', 'style', 'textarea', 'pre','code','a'],
            ignoreClass:"comment-content"
        },
        "HTML-CSS": {
            availableFonts: ["STIX","TeX"],
            showMathMenu: false
        }
    });
    MathJax.Hub.Queue(["Typeset",MathJax.Hub]);
    </script>
    <script src="https://cdn.bootcss.com/mathjax/2.7.0/MathJax.js?config=TeX-AMS-MML_HTMLorMML"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/3.0.1/github-markdown.min.css">
    <style>
        .markdown-body {
            width: 100%;
            margin-top: 45px;
            padding: 0px;
        }

        @media (max-width: 767px) {
            .markdown-body {
                padding: 0px;
            }
        }
    </style>
    <link rel="stylesheet" href="{{ URL::asset('css/articles/article_header_box.css') }}">
@stop

@section('extension_metas')
    <title>Article - {{ $article->title }}</title>
@stop

@section('main_content')
    <div class="margin-left-15 margin-right-15"> 
        <div class="header_box">
            <div class="clear" style="padding: 25px"></div>
            <div class="container">
                <span class="title">
                    <a href="{{url('articles', $article->id)}}">{{ $article->title }}</a>
                </span>
                <span class="tagbox">
                    <a class="category">Category: {{ $article->category }}</a>
                </span>
            </div>
            <div class="clear" style="padding: 12.5px"></div>
            <div class="container">
                <span class="tagbox">
                    <a class="timetag">Publish at: {{ $article->published_at }}</a>
                </span>
                <span class="tagbox">
                    <a class="timetag">Update at: {{ $article->updated_at }}</a>
                </span>
            </div>
        </div>
        <div class="markdown-body">
            {!! $content !!}
        </div>
    </div>
@stop
