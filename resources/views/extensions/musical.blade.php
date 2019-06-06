@extends('layouts.mainview')

@section('extension_files')
    <link rel="stylesheet" href="{{ URL::asset('css/extension/musical.css') }}">
    <script src="https://cdn.bootcss.com/audiojs/1.0.1/audio.min.js"></script>
    <script>
        setTimeout(function () {
            audiojs.events.ready(function() {
                var aAudioDomList = document.getElementsByTagName('audio')
                for(var i of aAudioDomList) {
                //初始化 自定义样式
                audiojs.create(i, {
                    css: false,
                    createPlayer: {
                    markup: false,
                    playPauseClass: 'play-pauseZ',
                    scrubberClass: 'scrubberZ',
                    progressClass: 'progressZ',
                    loaderClass: 'loadedZ',
                    timeClass: 'timeZ',
                    durationClass: 'durationZ',
                    playedClass: 'playedZ',
                    errorMessageClass: 'error-messageZ',
                    playingClass: 'playingZ',
                    loadingClass: 'loadingZ',
                    errorClass: 'errorZ'
                    }
                })
                }
            })
            //保存正在播放音乐的序号
            var sAudioPlayingIndex = '-1'
            for(var i in audiojs.instances){
                //去看了源码 没有找到点击播放按钮的回调函数 只能重写playPause（监听音乐播放进度事件）
                //
                audiojs.instances[i].playPause = function () {
                // 原playPause事件 
                if (this.playing) this.pause();
                else this.play();
                // 有正在播放的音乐序号与sAudioPlayingIndex所保存的不同 则暂停音乐
                var sId = this.wrapper.id.split('audiojs_wrapper')[1]
                if (this.playing === true && sAudioPlayingIndex !== sId) {
                    sAudioPlayingIndex = sId
                    for(var j in audiojs.instances){
                    if (j.split('audiojs')[1] != sAudioPlayingIndex) {
                        audiojs.instances[j].pause()
                    }
                    }
                }
                }
            }
            }, 2000);
    </script>
@stop

@section('extension_metas')
    <title>Musical</title>
@stop

@section('main_content')
<div class="musical_index_body"> 
        <div style="height: 50px; width:100%;"></div>
        @foreach($musics as $mmusic)
            <div class="index_box">
                <div class="title">
                    <a>{{ $mmusic->name }}</a>
                </div>
                <div class="tagbox">
                    <a class="artist">Artist: {{ $mmusic->artist }}&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Track: {{ $mmusic->track }}</a>
                </div>
                <div class="comment">
                    <a>{{ $mmusic->comment }}</a>
                </div>
                <div class="audiojsZ">
                    <audio src="{{ $mmusic->ref }}" preload="auto"></audio>
                    <div class="play-pauseZ">
                        <p class="playZ"></p>
                        <p class="pauseZ"></p>
                        <p class="loadingZ"></p>
                        <p class="errorZ"></p>
                    </div>
                    <div class="scrubberZ">
                        <div class="progressZ"></div>
                        <div class="loadedZ"></div>
                    </div>
                    <div class="timeZ" hidden>
                        <em class="playedZ" hidden>00:00</em>/<strong class="durationZ" hidden>00:00</strong>
                    </div>
                    <div class="error-messageZ"></div>
                </div>
            </div>
        @endforeach
    </div>
@stop
