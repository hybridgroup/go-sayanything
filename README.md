# sayanything

Go package wrappper for Text To Speech (TTS).

Supports the following local TTS engines:

- [Piper TTS](https://github.com/OHF-Voice/piper1-gpl)
- [Software Automatic Mouth (SAM)](https://github.com/deadprogram/sam)

It also supports the following cloud based TTS engines:

- [Google Cloud TTS](https://cloud.google.com/text-to-speech) 
    NOTE: has not been used or tested in quite a while.


## How to build

```
go build -o sayanything .
```

## How to run

### Piper


### SAM


### Google Cloud TTS

```
./sayanything -k="/path/to/key.json" -l="es-ES" -voice="es-ES-Neural2-E" "¡Hola amigo! ¿Cómo estás?"
```
